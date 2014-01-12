// Package clipboard provides a simple API for handling data exchange between
// the Go client and system clipboard.
// This package currently only supports small data strings for PRIMARY and
// CLIPBOARD selections.
//
// Useful things to read:
// Intro to X selections: http://www.jwz.org/doc/x-cut-and-paste.html
// Technical ICCCM manual: http://www.x.org/releases/X11R7.6/doc/xorg-docs/specs/ICCCM/icccm.html
package clipboard

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/icccm"
	"log"
)

const debug debugging = false

type debugging bool

func (d debugging) Printf(format string, args ...interface{}) {
	if d {
		log.Printf(format, args...)
	}
}

var (
	conn     *xgb.Conn
	connUtil *xgbutil.XUtil
	win      xproto.Window
	setup    *xproto.SetupInfo

	atomClipboard xproto.Atom
	atomTargets   xproto.Atom
	atomUtf8      xproto.Atom
	atomMultiple  xproto.Atom
	atomTimestamp xproto.Atom

	atoms = map[string]*xproto.Atom{
		"CLIPBOARD":   &atomClipboard,
		"TARGETS":     &atomTargets,
		"UTF8_STRING": &atomUtf8,
		"MULTIPLE":    &atomMultiple,
		"TIMESTAMP":   &atomTimestamp,
	}

	// Time when selection is acquired
	selectionTime xproto.Timestamp
)

// Performs initialization for clipboard to work. Not using init() to prevent
// unnecessary invocations.
func Setup() (err error) {
	if conn, err = xgb.NewConn(); err != nil {
		return
	}
	if connUtil, err = xgbutil.NewConn(); err != nil {
		return
	}

	setup = xproto.Setup(conn)          // Get setup info from X connection
	screen := setup.DefaultScreen(conn) // Get info from default screen

	// Generate window id
	if win, err = xproto.NewWindowId(conn); err != nil {
		return
	}

	if err = initAtoms(); err != nil {
		return
	}

	// Create dummy window for capturing Selection events
	xproto.CreateWindow(conn, 0, win, screen.Root, 0, 0, 1, 1, 0,
		xproto.WindowClassCopyFromParent, screen.RootVisual,
		xproto.CwEventMask,
		[]uint32{xproto.EventMaskPropertyChange})

	return
}

// Initializes all required atom identifiers that are used in this package
func initAtoms() (err error) {
	for name, atom := range atoms {
		atomCookie := xproto.InternAtom(conn, true, uint16(len(name)), name)
		atomReply, err := atomCookie.Reply()
		if err != nil {
			return err
		}
		*atom = atomReply.Atom
	}
	return
}

// Makes data available for PRIMARY and CLIPBOARD selection requests. This
// function does not block. Requestor names are sent to the string channel and
// errors are sent to the error channel.
func Put(data []byte) (<-chan string, <-chan error) {
	creq := make(chan string, 1)
	cerr := make(chan error, 1)

	if conn == nil {
		cerr <- errors.New("clipboard: X connection is nil")
		return creq, cerr
	}

	var err error
	selectionTime, err = getCurrentTime()
	if err != nil {
		cerr <- err
		return creq, cerr
	}

	go put(data, xproto.AtomPrimary, creq, cerr)
	go put(data, atomClipboard, creq, cerr)

	return creq, cerr
}

// Clear clears clipboard data by giving up selection ownership.
func Clear() {
	xproto.SetSelectionOwner(conn, xproto.AtomNone, xproto.AtomPrimary, selectionTime)
	xproto.SetSelectionOwner(conn, xproto.AtomNone, atomClipboard, selectionTime)
}

func put(data []byte, selection xproto.Atom,
	creq chan<- string, cerr chan<- error) {

	// Assert selection ownership
	xproto.SetSelectionOwner(conn, win, selection, selectionTime)

	// Check if selection owner has changed to our window
	cookie := xproto.GetSelectionOwner(conn, selection)
	reply, err := cookie.Reply()
	if err != nil {
		cerr <- err
		return
	}
	if reply.Owner != win {
		cerr <- errors.New("clipboard: unable to become selection owner")
		return
	}

	// Check if data exceeds maximum length
	if len(data) > int(setup.MaximumRequestLength) {
		cerr <- fmt.Errorf("clipboard: selection data length (%d) "+
			"exceeds maximum request length (%d)",
			len(data), setup.MaximumRequestLength)
	}

	// Start the main event loop.
	for {
		// WaitForEvent either returns an event or an error and never both.
		// If both are nil, then something went wrong and the loop should be
		// halted.
		//
		// An error can only be seen here as a response to an unchecked
		// request.
		var e xgb.Event
		if e, err = conn.WaitForEvent(); err != nil {
			return // Error can be ignored?
		}
		if e == nil {
			panic("both event and error are nil")
		}

		switch e := e.(type) {
		case xproto.SelectionRequestEvent:
			handleRequest(e, data, creq)

		case xproto.SelectionClearEvent:
			// log.Printf("lost selection: %s\n", e)
			return

		default:
			log.Printf("unknown event: %s\n", e)
		}
	}
}

func handleRequest(req xproto.SelectionRequestEvent,
	data []byte, creq chan<- string) {

	// Build response
	res := xproto.SelectionNotifyEvent{
		Requestor: req.Requestor,
		Selection: req.Selection,
		Target:    req.Target,
		Property:  req.Property,
		Time:      req.Time,
	}

	// If requestor is obsolete, set target as property atom
	if res.Property == xproto.AtomNone {
		log.Printf("requestor is obsolete\n")
		res.Property = res.Target
	}

	// Refuse request if request timestamp is earlier than time of initial
	// ownership
	var converted bool
	if req.Time != xproto.TimeCurrentTime && req.Time < selectionTime {
		req.Property = xproto.AtomNone
		log.Printf("incorrect timestamp; ignoring request\n")
	} else {
		converted = handleResponse(res, data)
	}

	// Notify requestor by responding with a SelectNotify event
	xproto.SendEvent(conn, false, res.Requestor,
		xproto.EventMaskNoEvent,
		string(res.Bytes()))

	if converted {
		name, _ := icccm.WmNameGet(connUtil, req.Requestor)
		// Run in separate goroutine so it doesn't block if chan is full
		go func() { creq <- name }()
	}
}

// Sends response to selection request by changing a designated property on
// requestor's window. A boolean is returned, indicating whether selection data
// is sent.
func handleResponse(res xproto.SelectionNotifyEvent, data []byte) bool {
	switch res.Target {
	// TARGETS: send list of supported target atoms
	// Some apps send multiple identical TARGETS requests and then revert
	// to UTF8_STRING/STRING, so this might not be functioning correctly
	case atomTargets:
		buf := new(bytes.Buffer)
		// Supported target atoms
		targets := []xproto.Atom{
			xproto.AtomString,
			atomUtf8,
			atomTargets,
			atomTimestamp,
		}
		for _, atom := range targets {
			binary.Write(buf, binary.LittleEndian, uint32(atom))
		}
		xproto.ChangeProperty(conn, xproto.PropModeReplace,
			res.Requestor, res.Property, xproto.AtomAtom,
			32, uint32(len(targets)), buf.Bytes())

	// TODO: MULTIPLE target case required?

	// TIMESTAMP: return timestamp used to acquire selection ownership
	case atomTimestamp:
		timestamp := make([]byte, 4)
		xgb.Put32(timestamp, uint32(selectionTime))
		xproto.ChangeProperty(conn, xproto.PropModeReplace,
			res.Requestor, res.Property, xproto.AtomInteger,
			32, 1, timestamp)

	// STRING, UTF8_STRING: send selection data
	case xproto.AtomString, atomUtf8:
		xproto.ChangeProperty(conn, xproto.PropModeReplace,
			res.Requestor, res.Property, res.Target,
			8, uint32(len(data)), data)

		return true

		// TODO: watch PropertyNotify event on requestor window to
		// verify data is transferred successfully (property must be deleted)

	// Unknown atom target: refuse request by setting Property to None
	default:
		log.Printf("request with unknown target: %s\n", res)
		res.Property = xproto.AtomNone
	}

	return false
}

// Returns time of X server
func getCurrentTime() (time xproto.Timestamp, err error) {
	// Perfrom a zero-length append to a property of our dummy window
	xproto.ChangeProperty(conn, xproto.PropModeAppend, win, xproto.AtomWmName,
		xproto.AtomString, 8, 0, []byte{})

	// Retrieve timestamp from PropertyNotify event
	for {
		var e xgb.Event
		if e, err = conn.WaitForEvent(); err != nil {
			return
		}
		if e == nil {
			panic("both event and error are nil")
		}
		if e, ok := e.(xproto.PropertyNotifyEvent); ok {
			return e.Time, nil
		}
	}
	return
}
