// API for data exchange with X11 clipboard (PRIMARY and CLIPBOARD)
// Global documentation: http://www.jwz.org/doc/x-cut-and-paste.html
// Technical documentation: http://tronche.com/gui/x/icccm/sec-2.html
package clipboard

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

var (
	conn  *xgb.Conn
	win   xproto.Window
	setup *xproto.SetupInfo
)

// Performs initialization for clipboard to work. Not using init() to prevent
// unnecessary invocations.
func Setup() (err error) {
	if conn, err = xgb.NewConn(); err != nil {
		return
	}

	setup = xproto.Setup(conn)          // Get setup info from X connection
	screen := setup.DefaultScreen(conn) // Get info from default screen

	// Generate window id
	if win, err = xproto.NewWindowId(conn); err != nil {
		return
	}

	// Create dummy window for capturing Selection events
	xproto.CreateWindow(conn, 0, win, screen.Root, 0, 0, 1, 1,
		0,
		xproto.WindowClassCopyFromParent, screen.RootVisual,
		xproto.CwEventMask,
		[]uint32{xproto.EventMaskPropertyChange})

	return
}

func Put(data []byte) (err error) {
	if conn == nil {
		return errors.New("no X connection, did you run clipboard.Setup() beforehand?")
	}

	fmt.Println("timestamps:")
	fmt.Println(xproto.Timestamp(0))
	fmt.Println(xproto.TimeCurrentTime)

	// Assert selection ownership
	var selection xproto.Atom = xproto.AtomPrimary
	time, err := getCurrentTime()
	if err != nil {
		return err
	}
	// time is 0
	xproto.SetSelectionOwner(conn, win, selection, time)

	// Check if selection owner has changed to our window
	cookie := xproto.GetSelectionOwner(conn, selection)
	reply, _ := cookie.Reply()
	// TODO: change panic to err
	if reply.Owner != win {
		panic("unable to become selection owner")
	}

	// Check if data exceeds maximum length
	if len(data) > int(setup.MaximumRequestLength) {
		return errors.New("selection data is too large") // add size
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
		e, err = conn.WaitForEvent()
		if e == nil && err == nil {
			fmt.Println("Both event and error are nil. Exiting...")
			// TODO: return err
			return
		}

		// type SelectionRequestEvent struct {
		//     Sequence uint16
		//     // padding: 1 bytes
		//     Time      Timestamp
		//     Owner     Window
		//     Requestor Window
		//     Selection Atom
		//     Target    Atom
		//     Property  Atom
		// }
		req, ok := e.(xproto.SelectionRequestEvent)
		if !ok {
			continue
		}
		if err != nil {
			panic(err)
		}

		// type SelectionNotifyEvent struct {
		//     Sequence uint16
		//     // padding: 1 bytes
		//     Time      Timestamp
		//     Requestor Window
		//     Selection Atom
		//     Target    Atom
		//     Property  Atom
		// }
		res := xproto.SelectionNotifyEvent{
			// sequence?
			Time:      req.Time,
			Requestor: req.Requestor,
			Selection: req.Selection,
			Target:    req.Target,
		}

		// Check for obsolete requestor
		if req.Property == xproto.AtomNone {
			res.Property = req.Target
		} else {
			res.Property = req.Property
		}

		fmt.Println(res)

		// if req.Target == xproto.AtomString {

		xproto.ChangeProperty(conn, xproto.PropModeReplace,
			res.Requestor, res.Property, xproto.AtomString,
			8,
			uint32(len(data)), data)

		fmt.Println(res.Bytes())
		xproto.SendEvent(conn, false, res.Requestor,
			xproto.EventMaskNoEvent,
			string(res.Bytes()))

		// }

		// check timestamp window => prop None

		// property

		fmt.Println(req)
	}
}

func getCurrentTime() (time xproto.Timestamp, err error) {
	// Perfrom a zero-length append to a property of our dummy window
	xproto.ChangeProperty(conn, xproto.PropModeAppend, win, xproto.AtomWmName,
		xproto.AtomString, 8, 0, []byte{})

	// Retrieve X' timestamp from PropertyNotify event
	for {
		var e xgb.Event
		if e, err = conn.WaitForEvent(); err != nil {
			return
		}

		if e, ok := e.(xproto.PropertyNotifyEvent); ok {
			return e.Time, nil
		}
	}

	return
}
