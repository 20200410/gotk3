/*
 * Copyright (c) 2013 Conformal Systems <info@conformal.com>
 *
 * This file originated from: http://opensource.conformal.com/
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

/*
Go bindings for GTK+ 3.  Supports version 3.8 and later.

Functions use the same names as the native C function calls, but use
CamelCase.  In cases where native GTK uses pointers to values to
simulate multiple return values, Go's native multiple return values
are used instead.  Whenever a native GTK call could return an
unexpected NULL pointer, an additonal error is returned in the Go
binding.

GTK's C API documentation can be very useful for understanding how the
functions in this package work and what each type is for.  This
documentation can be found at https://developer.gnome.org/gtk3/.

In addition to Go versions of the C GTK functions, every struct type
includes a function called Native(), taking itself as a receiver,
which returns the native C type or a pointer (in the case of
GObjects).  The returned C types are scoped to this gtk package and
must be converted to a local package before they can be used as
arguments to native GTK calls using cgo.

Memory management is handled in proper Go fashion, using runtime
finalizers to properly free memory when it is no longer needed.  Each
time a Go type is created with a pointer to a GObject, a reference is
added for Go, sinking the floating reference when necessary.  After
going out of scope and the next time Go's garbage collector is run, a
finalizer is run to remove Go's reference to the GObject.  When this
reference count hits zero (when neither Go nor GTK holds ownership)
the object will be freed internally by GTK.

It may be required to use these bindings with a patched Go, as GTK
seems to send signals on non-Go threads, causing the Go runtime to
promptly kill the process.  If you run into this error, try again
after patching Go with this patch:
https://codereview.appspot.com/10286046/.
*/
package gtk

// #cgo pkg-config: gtk+-3.0
// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"
import (
	"errors"
	"fmt"
	"github.com/conformal/gotk3/gdk"
	"github.com/conformal/gotk3/glib"
	"reflect"
	"runtime"
	"unsafe"
)

/*
 * Type conversions
 */

func gbool(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}
func gobool(b C.gboolean) bool {
	if b != 0 {
		return true
	}
	return false
}

// Wrapper function for TestBoolConvs since cgo can't be used with
// testing package
func testBoolConvs() error {
	b := gobool(gbool(true))
	if b != true {
		return errors.New("Unexpected bool conversion result")
	}

	cb := gbool(gobool(C.gboolean(0)))
	if cb != C.gboolean(0) {
		return errors.New("Unexpected bool conversion result")
	}

	return nil
}

/*
 * Unexported vars
 */

var nilPtrErr = errors.New("cgo returned unexpected nil pointer")

/*
 * Constants
 */

type Align int

const (
	ALIGN_FILL Align = iota
	ALIGN_START
	ALIGN_END
	ALIGN_CENTER
)

type ButtonsType int

const (
	BUTTONS_NONE ButtonsType = iota
	BUTTONS_OK
	BUTTONS_CLOSE
	BUTTONS_CANCEL
	BUTTONS_YES_NO
	BUTTONS_OK_CANCEL
)

type DialogFlags int

const (
	DIALOG_MODAL DialogFlags = 1 << iota
	DIALOG_DESTROY_WITH_PARENT
)

type EntryIconPosition int

const (
	ENTRY_ICON_PRIMARY EntryIconPosition = iota
	ENTRY_ICON_SECONDARY
)

type IconSize int

const (
	ICON_SIZE_INVALID IconSize = iota
	ICON_SIZE_MENU
	ICON_SIZE_SMALL_TOOLBAR
	ICON_SIZE_LARGE_TOOLBAR
	ICON_SIZE_BUTTON
	ICON_SIZE_DND
	ICON_SIZE_DIALOG
)

type ImageType int

const (
	IMAGE_EMPTY ImageType = iota
	IMAGE_PIXBUF
	IMAGE_STOCK
	IMAGE_ICON_SET
	IMAGE_ANIMATION
	IMAGE_ICON_NAME
	IMAGE_GICON
)

type InputHints int

const (
	INPUT_HINT_NONE       InputHints = 0
	INPUT_HINT_SPELLCHECK            = 1 << (iota - 1)
	INPUT_HINT_NO_SPELLCHECK
	INPUT_HINT_WORD_COMPLETION
	INPUT_HINT_LOWERCASE
	INPUT_HINT_UPPERCASE_CHARS
	INPUT_HINT_UPPERCASE_WORDS
	INPUT_HINT_UPPERCASE_SENTENCES
	INPUT_HINT_INHIBIT_OSK
)

type InputPurpose int

const (
	INPUT_PURPOSE_FREE_FORM InputPurpose = iota
	INPUT_PURPOSE_ALPHA
	INPUT_PURPOSE_DIGITS
	INPUT_PURPOSE_NUMBER
	INPUT_PURPOSE_PHONE
	INPUT_PURPOSE_URL
	INPUT_PURPOSE_EMAIL
	INPUT_PURPOSE_NAME
	INPUT_PURPOSE_PASSWORD
	INPUT_PURPOSE_PIN
)

type MessageType int

const (
	MESSAGE_INFO MessageType = iota
	MESSAGE_WARNING
	MESSAGE_QUESTION
	MESSAGE_ERROR
	MESSAGE_OTHER
)

type Orientation int

const (
	ORIENTATION_HORIZONTAL Orientation = iota
	ORIENTATION_VERTICAL
)

type PackType int

const (
	PACK_START PackType = iota
	PACK_END
)

type PolicyType int

const (
	POLICY_ALWAYS PolicyType = iota
	POLICY_AUTOMATIC
	POLICY_NEVER
)

type PositionType int

const (
	POS_LEFT PositionType = iota
	POS_RIGHT
	POS_TOP
	POS_BOTTOM
)

type ReliefStyle int

const (
	RELIEF_NORMAL ReliefStyle = iota
	RELIEF_HALF
	RELIEF_NONE
)

type ResponseType int

const (
	RESPONSE_NONE ResponseType = -(iota + 1)
	RESPONSE_REJECT
	RESPONSE_ACCEPT
	RESPONSE_DELETE_EVENT
	RESPONSE_OK
	RESPONSE_CANCEL
	RESPONSE_CLOSE
	RESPONSE_YES
	RESPONSE_NO
	RESPONSE_APPLY
	RESPONSE_HELP
)

type Stock string

const (
	STOCK_ABOUT                         Stock = "gtk-about"
	STOCK_ADD                           Stock = "gtk-add"
	STOCK_APPLY                         Stock = "gtk-apply"
	STOCK_BOLD                          Stock = "gtk-bold"
	STOCK_CANCEL                        Stock = "gtk-cancel"
	STOCK_CAPS_LOCK_WARNING             Stock = "gtk-caps-lock-warning"
	STOCK_CDROM                         Stock = "gtk-cdrom"
	STOCK_CLEAR                         Stock = "gtk-clear"
	STOCK_CLOSE                         Stock = "gtk-close"
	STOCK_COLOR_PICKER                  Stock = "gtk-color-picker"
	STOCK_CONNECT                       Stock = "gtk-connect"
	STOCK_CONVERT                       Stock = "gtk-convert"
	STOCK_COPY                          Stock = "gtk-copy"
	STOCK_CUT                           Stock = "gtk-cut"
	STOCK_DELETE                        Stock = "gtk-delete"
	STOCK_DIALOG_AUTHENTICATION         Stock = "gtk-dialog-authentication"
	STOCK_DIALOG_INFO                   Stock = "gtk-dialog-info"
	STOCK_DIALOG_WARNING                Stock = "gtk-dialog-warning"
	STOCK_DIALOG_ERROR                  Stock = "gtk-dialog-error"
	STOCK_DIALOG_QUESTION               Stock = "gtk-dialog-question"
	STOCK_DIRECTORY                     Stock = "gtk-directory"
	STOCK_DISCARD                       Stock = "gtk-discard"
	STOCK_DISCONNECT                    Stock = "gtk-disconnect"
	STOCK_DND                           Stock = "gtk-dnd"
	STOCK_DND_MULTIPLE                  Stock = "gtk-dnd-multiple"
	STOCK_EDIT                          Stock = "gtk-edit"
	STOCK_EXECUTE                       Stock = "gtk-execute"
	STOCK_FILE                          Stock = "gtk-file"
	STOCK_FIND                          Stock = "gtk-find"
	STOCK_FIND_AND_REPLACE              Stock = "gtk-find-and-replace"
	STOCK_FLOPPY                        Stock = "gtk-floppy"
	STOCK_FULLSCREEN                    Stock = "gtk-fullscreen"
	STOCK_GOTO_BOTTOM                   Stock = "gtk-goto-bottom"
	STOCK_GOTO_FIRST                    Stock = "gtk-goto-first"
	STOCK_GOTO_LAST                     Stock = "gtk-goto-last"
	STOCK_GOTO_TOP                      Stock = "gtk-goto-top"
	STOCK_GO_BACK                       Stock = "gtk-go-back"
	STOCK_GO_DOWN                       Stock = "gtk-go-down"
	STOCK_GO_FORWARD                    Stock = "gtk-go-forward"
	STOCK_GO_UP                         Stock = "gtk-go-up"
	STOCK_HARDDISK                      Stock = "gtk-harddisk"
	STOCK_HELP                          Stock = "gtk-help"
	STOCK_HOME                          Stock = "gtk-home"
	STOCK_INDEX                         Stock = "gtk-index"
	STOCK_INDENT                        Stock = "gtk-indent"
	STOCK_INFO                          Stock = "gtk-info"
	STOCK_ITALIC                        Stock = "gtk-italic"
	STOCK_JUMP_TO                       Stock = "gtk-jump-to"
	STOCK_JUSTIFY_CENTER                Stock = "gtk-justify-center"
	STOCK_JUSTIFY_FILL                  Stock = "gtk-justify-fill"
	STOCK_JUSTIFY_LEFT                  Stock = "gtk-justify-left"
	STOCK_JUSTIFY_RIGHT                 Stock = "gtk-justify-right"
	STOCK_LEAVE_FULLSCREEN              Stock = "gtk-leave-fullscreen"
	STOCK_MISSING_IMAGE                 Stock = "gtk-missing-image"
	STOCK_MEDIA_FORWARD                 Stock = "gtk-media-forward"
	STOCK_MEDIA_NEXT                    Stock = "gtk-media-next"
	STOCK_MEDIA_PAUSE                   Stock = "gtk-media-pause"
	STOCK_MEDIA_PLAY                    Stock = "gtk-media-play"
	STOCK_MEDIA_PREVIOUS                Stock = "gtk-media-previous"
	STOCK_MEDIA_RECORD                  Stock = "gtk-media-record"
	STOCK_MEDIA_REWIND                  Stock = "gtk-media-rewind"
	STOCK_MEDIA_STOP                    Stock = "gtk-media-stop"
	STOCK_NETWORK                       Stock = "gtk-network"
	STOCK_NEW                           Stock = "gtk-new"
	STOCK_NO                            Stock = "gtk-no"
	STOCK_OK                            Stock = "gtk-ok"
	STOCK_OPEN                          Stock = "gtk-open"
	STOCK_ORIENTATION_PORTRAIT          Stock = "gtk-orientation-portrait"
	STOCK_ORIENTATION_LANDSCAPE         Stock = "gtk-orientation-landscape"
	STOCK_ORIENTATION_REVERSE_LANDSCAPE Stock = "gtk-orientation-reverse-landscape"
	STOCK_ORIENTATION_REVERSE_PORTRAIT  Stock = "gtk-orientation-reverse-portrait"
	STOCK_PAGE_SETUP                    Stock = "gtk-page-setup"
	STOCK_PASTE                         Stock = "gtk-paste"
	STOCK_PREFERENCES                   Stock = "gtk-preferences"
	STOCK_PRINT                         Stock = "gtk-print"
	STOCK_PRINT_ERROR                   Stock = "gtk-print-error"
	STOCK_PRINT_PAUSED                  Stock = "gtk-print-paused"
	STOCK_PRINT_PREVIEW                 Stock = "gtk-print-preview"
	STOCK_PRINT_REPORT                  Stock = "gtk-print-report"
	STOCK_PRINT_WARNING                 Stock = "gtk-print-warning"
	STOCK_PROPERTIES                    Stock = "gtk-properties"
	STOCK_QUIT                          Stock = "gtk-quit"
	STOCK_REDO                          Stock = "gtk-redo"
	STOCK_REFRESH                       Stock = "gtk-refresh"
	STOCK_REMOVE                        Stock = "gtk-remove"
	STOCK_REVERT_TO_SAVED               Stock = "gtk-revert-to-saved"
	STOCK_SAVE                          Stock = "gtk-save"
	STOCK_SAVE_AS                       Stock = "gtk-save-as"
	STOCK_SELECT_ALL                    Stock = "gtk-select-all"
	STOCK_SELECT_COLOR                  Stock = "gtk-select-color"
	STOCK_SELECT_FONT                   Stock = "gtk-select-font"
	STOCK_SORT_ASCENDING                Stock = "gtk-sort-ascending"
	STOCK_SORT_DESCENDING               Stock = "gtk-sort-descending"
	STOCK_SPELL_CHECK                   Stock = "gtk-spell-check"
	STOCK_STOP                          Stock = "gtk-stop"
	STOCK_STRIKETHROUGH                 Stock = "gtk-strikethrough"
	STOCK_UNDELETE                      Stock = "gtk-undelete"
	STOCK_UNDERLINE                     Stock = "gtk-underline"
	STOCK_UNDO                          Stock = "gtk-undo"
	STOCK_UNINDENT                      Stock = "gtk-unindent"
	STOCK_YES                           Stock = "gtk-yes"
	STOCK_ZOOM_100                      Stock = "gtk-zoom-100"
	STOCK_ZOOM_FIT                      Stock = "gtk-zoom-fit"
	STOCK_ZOOM_IN                       Stock = "gtk-zoom-in"
	STOCK_ZOOM_OUT                      Stock = "gtk-zoom-out"
)

type TreeModelFlags int

const (
	TREE_MODEL_ITERS_PERSIST TreeModelFlags = 1 << iota
	TREE_MODEL_LIST_ONLY
)

type WindowPosition int

const (
	WIN_POS_NONE WindowPosition = iota
	WIN_POS_CENTER
	WIN_POS_MOUSE
	WIN_POS_CENTER_ALWAYS
	WIN_POS_CENTER_ON_PARENT
)

type WindowType int

const (
	WINDOW_TOPLEVEL WindowType = iota
	WINDOW_POPUP
)

// Wrapper function for TestConsts since cgo can't be used with
// testing package
func testConsts() error {
	tests := []struct {
		GoConst  interface{}
		GtkConst interface{}
	}{
		{ALIGN_FILL, C.GTK_ALIGN_FILL},
		{ALIGN_START, C.GTK_ALIGN_START},
		{ALIGN_END, C.GTK_ALIGN_END},
		{ALIGN_CENTER, C.GTK_ALIGN_CENTER},

		{BUTTONS_NONE, C.GTK_BUTTONS_NONE},
		{BUTTONS_OK, C.GTK_BUTTONS_OK},
		{BUTTONS_CLOSE, C.GTK_BUTTONS_CLOSE},
		{BUTTONS_CANCEL, C.GTK_BUTTONS_CANCEL},
		{BUTTONS_YES_NO, C.GTK_BUTTONS_YES_NO},
		{BUTTONS_OK_CANCEL, C.GTK_BUTTONS_OK_CANCEL},

		{DIALOG_MODAL, C.GTK_DIALOG_MODAL},
		{DIALOG_DESTROY_WITH_PARENT, C.GTK_DIALOG_DESTROY_WITH_PARENT},

		{ENTRY_ICON_PRIMARY, C.GTK_ENTRY_ICON_PRIMARY},
		{ENTRY_ICON_SECONDARY, C.GTK_ENTRY_ICON_SECONDARY},

		{IMAGE_EMPTY, C.GTK_IMAGE_EMPTY},
		{IMAGE_PIXBUF, C.GTK_IMAGE_PIXBUF},
		{IMAGE_STOCK, C.GTK_IMAGE_STOCK},
		{IMAGE_ICON_SET, C.GTK_IMAGE_ICON_SET},
		{IMAGE_ANIMATION, C.GTK_IMAGE_ANIMATION},
		{IMAGE_ICON_NAME, C.GTK_IMAGE_ICON_NAME},
		{IMAGE_GICON, C.GTK_IMAGE_GICON},

		{INPUT_HINT_NONE, C.GTK_INPUT_HINT_NONE},
		{INPUT_HINT_SPELLCHECK, C.GTK_INPUT_HINT_SPELLCHECK},
		{INPUT_HINT_NO_SPELLCHECK, C.GTK_INPUT_HINT_NO_SPELLCHECK},
		{INPUT_HINT_WORD_COMPLETION, C.GTK_INPUT_HINT_WORD_COMPLETION},
		{INPUT_HINT_LOWERCASE, C.GTK_INPUT_HINT_LOWERCASE},
		{INPUT_HINT_UPPERCASE_CHARS, C.GTK_INPUT_HINT_UPPERCASE_CHARS},
		{INPUT_HINT_UPPERCASE_WORDS, C.GTK_INPUT_HINT_UPPERCASE_WORDS},
		{INPUT_HINT_UPPERCASE_SENTENCES, C.GTK_INPUT_HINT_UPPERCASE_SENTENCES},
		{INPUT_HINT_INHIBIT_OSK, C.GTK_INPUT_HINT_INHIBIT_OSK},

		{INPUT_PURPOSE_FREE_FORM, C.GTK_INPUT_PURPOSE_FREE_FORM},
		{INPUT_PURPOSE_ALPHA, C.GTK_INPUT_PURPOSE_ALPHA},
		{INPUT_PURPOSE_DIGITS, C.GTK_INPUT_PURPOSE_DIGITS},
		{INPUT_PURPOSE_NUMBER, C.GTK_INPUT_PURPOSE_NUMBER},
		{INPUT_PURPOSE_PHONE, C.GTK_INPUT_PURPOSE_PHONE},
		{INPUT_PURPOSE_URL, C.GTK_INPUT_PURPOSE_URL},
		{INPUT_PURPOSE_EMAIL, C.GTK_INPUT_PURPOSE_EMAIL},
		{INPUT_PURPOSE_NAME, C.GTK_INPUT_PURPOSE_NAME},
		{INPUT_PURPOSE_PASSWORD, C.GTK_INPUT_PURPOSE_PASSWORD},
		{INPUT_PURPOSE_PIN, C.GTK_INPUT_PURPOSE_PIN},

		{MESSAGE_INFO, C.GTK_MESSAGE_INFO},
		{MESSAGE_WARNING, C.GTK_MESSAGE_WARNING},
		{MESSAGE_QUESTION, C.GTK_MESSAGE_QUESTION},
		{MESSAGE_ERROR, C.GTK_MESSAGE_ERROR},
		{MESSAGE_OTHER, C.GTK_MESSAGE_OTHER},

		{ORIENTATION_HORIZONTAL, C.GTK_ORIENTATION_HORIZONTAL},
		{ORIENTATION_VERTICAL, C.GTK_ORIENTATION_VERTICAL},

		{PACK_START, C.GTK_PACK_START},
		{PACK_END, C.GTK_PACK_END},

		{POLICY_ALWAYS, C.GTK_POLICY_ALWAYS},
		{POLICY_AUTOMATIC, C.GTK_POLICY_AUTOMATIC},
		{POLICY_NEVER, C.GTK_POLICY_NEVER},

		{POS_LEFT, C.GTK_POS_LEFT},
		{POS_RIGHT, C.GTK_POS_RIGHT},
		{POS_TOP, C.GTK_POS_TOP},
		{POS_BOTTOM, C.GTK_POS_BOTTOM},

		{RELIEF_NORMAL, C.GTK_RELIEF_NORMAL},
		{RELIEF_HALF, C.GTK_RELIEF_HALF},
		{RELIEF_NONE, C.GTK_RELIEF_NONE},

		{RESPONSE_NONE, C.GTK_RESPONSE_NONE},
		{RESPONSE_REJECT, C.GTK_RESPONSE_REJECT},
		{RESPONSE_ACCEPT, C.GTK_RESPONSE_ACCEPT},
		{RESPONSE_DELETE_EVENT, C.GTK_RESPONSE_DELETE_EVENT},
		{RESPONSE_OK, C.GTK_RESPONSE_OK},
		{RESPONSE_CANCEL, C.GTK_RESPONSE_CANCEL},
		{RESPONSE_CLOSE, C.GTK_RESPONSE_CLOSE},
		{RESPONSE_YES, C.GTK_RESPONSE_YES},
		{RESPONSE_NO, C.GTK_RESPONSE_NO},
		{RESPONSE_APPLY, C.GTK_RESPONSE_APPLY},
		{RESPONSE_HELP, C.GTK_RESPONSE_HELP},

		{TREE_MODEL_ITERS_PERSIST, C.GTK_TREE_MODEL_ITERS_PERSIST},
		{TREE_MODEL_LIST_ONLY, C.GTK_TREE_MODEL_LIST_ONLY},

		{WIN_POS_NONE, C.GTK_WIN_POS_NONE},
		{WIN_POS_CENTER, C.GTK_WIN_POS_CENTER},
		{WIN_POS_MOUSE, C.GTK_WIN_POS_MOUSE},
		{WIN_POS_CENTER_ALWAYS, C.GTK_WIN_POS_CENTER_ALWAYS},
		{WIN_POS_CENTER_ON_PARENT, C.GTK_WIN_POS_CENTER_ON_PARENT},

		{WINDOW_TOPLEVEL, C.GTK_WINDOW_TOPLEVEL},
		{WINDOW_POPUP, C.GTK_WINDOW_POPUP},
	}
	for i, test := range tests {
		v := reflect.ValueOf(test.GoConst)
		iv := int(v.Int())
		if iv != test.GtkConst.(int) {
			return fmt.Errorf("Constant mismatch %d: %d != %d", i, iv, test.GtkConst.(int))
		}
	}
	return nil
}

/*
 * Init and main event loop
 */

/*
Init() must be called before any other GTK calls and is used to
initialize everything necessary.

In addition to setting up GTK for usage, a pointer to a slice of
strings may be passed in to parse standard GTK command line arguments.
args will be modified to remove any flags that were handled.
Alternatively, nil may be passed in to not perform any command line
parsing.
*/
func Init(args *[]string) {
	if args != nil {
		argc := C.int(len(*args))
		argv := make([]*C.char, argc)
		for i, arg := range *args {
			argv[i] = C.CString(arg)
		}
		C.gtk_init((*C.int)(unsafe.Pointer(&argc)),
			(***C.char)(unsafe.Pointer(&argv)))
		unhandled := make([]string, argc)
		for i := 0; i < int(argc); i++ {
			unhandled[i] = C.GoString(argv[i])
			C.free(unsafe.Pointer(argv[i]))
		}
		*args = unhandled
	} else {
		C.gtk_init(nil, nil)
	}
}

// Main() runs the GTK main loop, blocking until MainQuit() is called.
func Main() {
	C.gtk_main()
}

// MainQuit() is used to terminate the GTK main loop (started by Main()).
func MainQuit() {
	C.gtk_main_quit()
}

/*
 * GtkAdjustment
 */

// Adjustment is a type representing an adjustable bounded value
type Adjustment struct {
	glib.InitiallyUnowned
}

func (v *Adjustment) Native() *C.GtkAdjustment {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkAdjustment(p)
}

/*
 * GtkBin
 */

// Bin is a container that is restricted to holding a maximum of one
// child.
type Bin struct {
	Container
}

func (v *Bin) Native() *C.GtkBin {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkBin(p)
}

func (v *Bin) GetChild() (*Widget, error) {
	c := C.gtk_bin_get_child(v.Native())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	w := &Widget{glib.InitiallyUnowned{obj}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return w, nil
}

/*
 * GtkButton
 */

// Button is a widget that emits a signal whenever it is clicked.
// A Button is rendered using the contents of the Bin it embeds.
type Button struct {
	Bin
}

func (v *Button) Native() *C.GtkButton {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkButton(p)
}

func ButtonNew() (*Button, error) {
	c := C.gtk_button_new()
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	b := &Button{Bin{Container{Widget{glib.InitiallyUnowned{obj}}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return b, nil
}

func ButtonNewWithLabel(label string) (*Button, error) {
	cstr := C.CString(label)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_button_new_with_label((*C.gchar)(cstr))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	b := &Button{Bin{Container{Widget{glib.InitiallyUnowned{obj}}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return b, nil
}

func ButtonNewFromStock(stock Stock) (*Button, error) {
	cstr := C.CString(string(stock))
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_button_new_from_stock((*C.gchar)(cstr))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	b := &Button{Bin{Container{Widget{glib.InitiallyUnowned{obj}}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return b, nil
}

func ButtonNewWithMnemonic(label string) (*Button, error) {
	cstr := C.CString(label)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_button_new_with_mnemonic((*C.gchar)(cstr))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	b := &Button{Bin{Container{Widget{glib.InitiallyUnowned{obj}}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return b, nil
}

func (v *Button) Clicked() {
	C.gtk_button_clicked(v.Native())
}

func (v *Button) SetRelief(newStyle ReliefStyle) {
	C.gtk_button_set_relief(v.Native(), C.GtkReliefStyle(newStyle))
}

func (v *Button) GetRelief() ReliefStyle {
	c := C.gtk_button_get_relief(v.Native())
	return ReliefStyle(c)
}

func (v *Button) SetLabel(label string) {
	cstr := C.CString(label)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_button_set_label(v.Native(), (*C.gchar)(cstr))
}

func (v *Button) GetLabel() (string, error) {
	c := C.gtk_button_get_label(v.Native())
	if c == nil {
		return "", nilPtrErr
	}
	return C.GoString((*C.char)(c)), nil
}

func (v *Button) SetUseUnderline(useUnderline bool) {
	C.gtk_button_set_use_underline(v.Native(), gbool(useUnderline))
}

func (v *Button) GetUseUnderline() bool {
	c := C.gtk_button_get_use_underline(v.Native())
	return gobool(c)
}

func (v *Button) SetUseStock(useStock bool) {
	C.gtk_button_set_use_stock(v.Native(), gbool(useStock))
}

func (v *Button) GetUseStock() bool {
	c := C.gtk_button_get_use_stock(v.Native())
	return gobool(c)
}

func (v *Button) SetFocusOnClick(focusOnClick bool) {
	C.gtk_button_set_focus_on_click(v.Native(), gbool(focusOnClick))
}

func (v *Button) GetFocusOnClick() bool {
	c := C.gtk_button_get_focus_on_click(v.Native())
	return gobool(c)
}

func (v *Button) SetAlignment(xalign, yalign float32) {
	C.gtk_button_set_alignment(v.Native(), (C.gfloat)(xalign),
		(C.gfloat)(yalign))
}

func (v *Button) GetAlignment() (xalign, yalign float32) {
	var x, y C.gfloat
	C.gtk_button_get_alignment(v.Native(), &x, &y)
	return float32(x), float32(y)
}

func (v *Button) SetImage(image IWidget) {
	C.gtk_button_set_image(v.Native(), image.toWidget())
}

func (v *Button) GetImage() (*Widget, error) {
	c := C.gtk_button_get_image(v.Native())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	w := &Widget{glib.InitiallyUnowned{obj}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return w, nil
}

func (v *Button) SetImagePosition(position PositionType) {
	C.gtk_button_set_image_position(v.Native(), C.GtkPositionType(position))
}

func (v *Button) GetImagePosition() PositionType {
	c := C.gtk_button_get_image_position(v.Native())
	return PositionType(c)
}

func (v *Button) SetAlwaysShowImage(alwaysShow bool) {
	C.gtk_button_set_always_show_image(v.Native(), gbool(alwaysShow))
}

func (v *Button) GetAlwaysShowImage() bool {
	c := C.gtk_button_get_always_show_image(v.Native())
	return gobool(c)
}

func (v *Button) GetEventWindow() (*gdk.Window, error) {
	c := C.gtk_button_get_event_window(v.Native())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	w := &gdk.Window{obj}
	w.Ref()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return w, nil
}

/*
 * GtkBox
 */

type Box struct {
	Container
}

func (v *Box) Native() *C.GtkBox {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkBox(p)
}

func BoxNew(orientation Orientation, spacing int) (*Box, error) {
	c := C.gtk_box_new(C.GtkOrientation(orientation), C.gint(spacing))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	b := &Box{Container{Widget{glib.InitiallyUnowned{obj}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return b, nil
}

func (v *Box) PackStart(child IWidget, expand, fill bool, padding uint) {
	C.gtk_box_pack_start(v.Native(), child.toWidget(), gbool(expand),
		gbool(fill), C.guint(padding))
}

func (v *Box) PackEnd(child IWidget, expand, fill bool, padding uint) {
	C.gtk_box_pack_end(v.Native(), child.toWidget(), gbool(expand),
		gbool(fill), C.guint(padding))
}

func (v *Box) GetHomogeneous() bool {
	c := C.gtk_box_get_homogeneous(v.Native())
	return gobool(c)
}

func (v *Box) SetHomogeneous(homogeneous bool) {
	C.gtk_box_set_homogeneous(v.Native(), gbool(homogeneous))
}

func (v *Box) GetSpacing() int {
	c := C.gtk_box_get_spacing(v.Native())
	return int(c)
}

func (v *Box) SetSpacing(spacing int) {
	C.gtk_box_set_spacing(v.Native(), C.gint(spacing))
}

func (v *Box) ReorderChild(child IWidget, position int) {
	C.gtk_box_reorder_child(v.Native(), child.toWidget(), C.gint(position))
}

func (v *Box) QueryChildPacking(child IWidget) (expand, fill bool, padding uint, packType PackType) {
	var cexpand, cfill C.gboolean
	var cpadding C.guint
	var cpackType C.GtkPackType

	C.gtk_box_query_child_packing(v.Native(), child.toWidget(), &cexpand,
		&cfill, &cpadding, &cpackType)
	return gobool(cexpand), gobool(cfill), uint(cpadding), PackType(cpackType)
}

func (v *Box) SetChildPacking(child IWidget, expand, fill bool, padding uint, packType PackType) {
	C.gtk_box_set_child_packing(v.Native(), child.toWidget(), gbool(expand),
		gbool(fill), C.guint(padding), C.GtkPackType(packType))
}

/*
 * GtkCellLayout
 */

type CellLayout struct {
	*glib.Object
}

type ICellLayout interface {
	toCellLayout() *C.GtkCellLayout
}

func (v *CellLayout) Native() *C.GtkCellLayout {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkCellLayout(p)
}

func (v *CellLayout) toCellLayout() *C.GtkCellLayout {
	if v == nil {
		return nil
	}
	return v.Native()
}

func (v *CellLayout) PackStart(cell ICellRenderer, expand bool) {
	C.gtk_cell_layout_pack_start(v.Native(), cell.toCellRenderer(),
		gbool(expand))
}

func (v *CellLayout) AddAttribute(cell ICellRenderer, attribute string, column int) {
	cstr := C.CString(attribute)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_cell_layout_add_attribute(v.Native(), cell.toCellRenderer(),
		(*C.gchar)(cstr), C.gint(column))
}

/*
 * GtkCellRenderer
 */

type CellRenderer struct {
	glib.InitiallyUnowned
}

type ICellRenderer interface {
	toCellRenderer() *C.GtkCellRenderer
}

func (v *CellRenderer) Native() *C.GtkCellRenderer {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkCellRenderer(p)
}

func (v *CellRenderer) toCellRenderer() *C.GtkCellRenderer {
	if v == nil {
		return nil
	}
	return v.Native()
}

/*
 * GtkCellRendererText
 */

type CellRendererText struct {
	CellRenderer
}

func (v *CellRendererText) Native() *C.GtkCellRendererText {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkCellRendererText(p)
}

func (v *CellRendererText) toCellRenderer() *C.GtkCellRenderer {
	if v == nil {
		return nil
	}
	return v.CellRenderer.Native()
}

func CellRendererTextNew() (*CellRendererText, error) {
	c := C.gtk_cell_renderer_text_new()
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	crt := &CellRendererText{CellRenderer{glib.InitiallyUnowned{obj}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return crt, nil
}

/*
 * GtkClipboard
 */

type Clipboard struct {
	*glib.Object
}

func (v *Clipboard) Native() *C.GtkClipboard {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkClipboard(p)
}

func ClipboardGet(atom gdk.Atom) (*Clipboard, error) {
	c := C.gtk_clipboard_get(atom.Native())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	cb := &Clipboard{obj}
	obj.Ref()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return cb, nil
}

func ClipboardGetForDisplay(display *gdk.Display, atom gdk.Atom) (*Clipboard, error) {
	c := C.gtk_clipboard_get_for_display(display.Native(), atom.Native())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	cb := &Clipboard{obj}
	obj.Ref()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return cb, nil
}

func (v *Clipboard) SetText(text string) {
	cstr := C.CString(text)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_clipboard_set_text(v.Native(), (*C.gchar)(cstr),
		C.gint(len(text)))
}

/*
 * GtkComboBox
 */

type ComboBox struct {
	Bin

	// Interfaces
	CellLayout
}

func (v *ComboBox) Native() *C.GtkComboBox {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkComboBox(p)
}

func (v *ComboBox) toCellLayout() *C.GtkCellLayout {
	if v == nil {
		return nil
	}
	return C.toGtkCellLayout(unsafe.Pointer(v.GObject))
}

func ComboBoxNew() (*ComboBox, error) {
	c := C.gtk_combo_box_new()
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	cl := CellLayout{obj}
	cb := &ComboBox{Bin{Container{Widget{glib.InitiallyUnowned{obj}}}}, cl}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return cb, nil
}

func ComboBoxNewWithEntry() (*ComboBox, error) {
	c := C.gtk_combo_box_new_with_entry()
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	cl := CellLayout{obj}
	cb := &ComboBox{Bin{Container{Widget{glib.InitiallyUnowned{obj}}}}, cl}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return cb, nil
}

func ComboBoxNewWithModel(model ITreeModel) (*ComboBox, error) {
	c := C.gtk_combo_box_new_with_model(model.toTreeModel())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	cl := CellLayout{obj}
	cb := &ComboBox{Bin{Container{Widget{glib.InitiallyUnowned{obj}}}}, cl}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return cb, nil
}

func (v *ComboBox) GetActive() int {
	c := C.gtk_combo_box_get_active(v.Native())
	return int(c)
}

func (v *ComboBox) SetActive(index int) {
	C.gtk_combo_box_set_active(v.Native(), C.gint(index))
}

/*
 * GtkContainer
 */

type Container struct {
	Widget
}

func (v *Container) Native() *C.GtkContainer {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkContainer(p)
}

func (v *Container) Add(w IWidget) {
	C.gtk_container_add(v.Native(), w.toWidget())
}

func (v *Container) Remove(w IWidget) {
	C.gtk_container_remove(v.Native(), w.toWidget())
}

// Others...

/*
 * GtkDialog
 */

type Dialog struct {
	Window
}

func (v *Dialog) Native() *C.GtkDialog {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkDialog(p)
}

func DialogNew() (*Dialog, error) {
	c := C.gtk_dialog_new()
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	d := &Dialog{Window{Bin{Container{Widget{glib.InitiallyUnowned{obj}}}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return d, nil
}

func (v *Dialog) Run() int {
	c := C.gtk_dialog_run(v.Native())
	return int(c)
}

func (v *Dialog) Response(response ResponseType) {
	C.gtk_dialog_response(v.Native(), C.gint(response))
}

// text may be either the literal button text, or a StockID.
func (v *Dialog) AddButton(text string, id ResponseType) (*Button, error) {
	cstr := C.CString(text)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_dialog_add_button(v.Native(), (*C.gchar)(cstr), C.gint(id))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	b := &Button{Bin{Container{Widget{glib.InitiallyUnowned{obj}}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return b, nil
}

func (v *Dialog) AddActionWidget(child IWidget, id ResponseType) {
	C.gtk_dialog_add_action_widget(v.Native(), child.toWidget(), C.gint(id))
}

func (v *Dialog) SetDefaultResponse(id ResponseType) {
	C.gtk_dialog_set_default_response(v.Native(), C.gint(id))
}

func (v *Dialog) SetResponseSensitive(id ResponseType, setting bool) {
	C.gtk_dialog_set_response_sensitive(v.Native(), C.gint(id),
		gbool(setting))
}

func (v *Dialog) GetResponseForWidget(widget IWidget) ResponseType {
	c := C.gtk_dialog_get_response_for_widget(v.Native(), widget.toWidget())
	return ResponseType(c)
}

func (v *Dialog) GetWidgetForResponse(id ResponseType) (*Widget, error) {
	c := C.gtk_dialog_get_widget_for_response(v.Native(), C.gint(id))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	w := &Widget{glib.InitiallyUnowned{obj}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return w, nil
}

func (v *Dialog) GetActionArea() (*Widget, error) {
	c := C.gtk_dialog_get_action_area(v.Native())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	w := &Widget{glib.InitiallyUnowned{obj}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return w, nil
}

func (v *Dialog) GetContentArea() (*Box, error) {
	c := C.gtk_dialog_get_content_area(v.Native())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	b := &Box{Container{Widget{glib.InitiallyUnowned{obj}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return b, nil
}

// TODO(jrick)
/*
func (v *gdk.Screen) AlternativeDialogButtonOrder() bool {
	c := C.gtk_alternative_dialog_button_order(v.Native())
	return gobool(c)
}
*/

// TODO(jrick)
/*
func SetAlternativeButtonOrder(ids ...ResponseType) {
}
*/

/*
 * GtkEntry
 */

type Entry struct {
	Widget
}

func (v *Entry) Native() *C.GtkEntry {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkEntry(p)
}

func EntryNew() (*Entry, error) {
	c := C.gtk_entry_new()
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	e := &Entry{Widget{glib.InitiallyUnowned{obj}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return e, nil
}

func EntryNewWithBuffer(buffer *EntryBuffer) (*Entry, error) {
	c := C.gtk_entry_new_with_buffer(buffer.Native())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	e := &Entry{Widget{glib.InitiallyUnowned{obj}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return e, nil
}

func (v *Entry) GetBuffer() (*EntryBuffer, error) {
	c := C.gtk_entry_get_buffer(v.Native())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	e := &EntryBuffer{obj}
	obj.Ref()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return e, nil
}

func (v *Entry) SetBuffer(buffer *EntryBuffer) {
	C.gtk_entry_set_buffer(v.Native(), buffer.Native())
}

func (v *Entry) SetText(text string) {
	cstr := C.CString(text)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_entry_set_text(v.Native(), (*C.gchar)(cstr))
}

func (v *Entry) GetText() (string, error) {
	c := C.gtk_entry_get_text(v.Native())
	if c == nil {
		return "", nilPtrErr
	}
	return C.GoString((*C.char)(c)), nil
}

func (v *Entry) GetTextLength() uint16 {
	c := C.gtk_entry_get_text_length(v.Native())
	return uint16(c)
}

// TODO(jrick) GdkRectangle
func (v *Entry) GetTextArea() {
}

func (v *Entry) SetVisibility(visible bool) {
	C.gtk_entry_set_visibility(v.Native(), gbool(visible))
}

func (v *Entry) SetInvisibleChar(ch rune) {
	C.gtk_entry_set_invisible_char(v.Native(), C.gunichar(ch))
}

func (v *Entry) UnsetInvisibleChar() {
	C.gtk_entry_unset_invisible_char(v.Native())
}

func (v *Entry) SetMaxLength(len int) {
	C.gtk_entry_set_max_length(v.Native(), C.gint(len))
}

func (v *Entry) GetActivatesDefault() bool {
	c := C.gtk_entry_get_activates_default(v.Native())
	return gobool(c)
}

func (v *Entry) GetHasFrame() bool {
	c := C.gtk_entry_get_has_frame(v.Native())
	return gobool(c)
}

func (v *Entry) GetWidthChars() int {
	c := C.gtk_entry_get_width_chars(v.Native())
	return int(c)
}

func (v *Entry) SetActivatesDefault(setting bool) {
	C.gtk_entry_set_activates_default(v.Native(), gbool(setting))
}

func (v *Entry) SetHasFrame(setting bool) {
	C.gtk_entry_set_has_frame(v.Native(), gbool(setting))
}

func (v *Entry) SetWidthChars(nChars int) {
	C.gtk_entry_set_width_chars(v.Native(), C.gint(nChars))
}

func (v *Entry) GetInvisibleChar() rune {
	c := C.gtk_entry_get_invisible_char(v.Native())
	return rune(c)
}

func (v *Entry) SetAlignment(xalign float32) {
	C.gtk_entry_set_alignment(v.Native(), C.gfloat(xalign))
}

func (v *Entry) GetAlignment() float32 {
	c := C.gtk_entry_get_alignment(v.Native())
	return float32(c)
}

func (v *Entry) SetPlaceholderText(text string) {
	cstr := C.CString(text)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_entry_set_placeholder_text(v.Native(), (*C.gchar)(cstr))
}

func (v *Entry) GetPlaceholderText() (string, error) {
	c := C.gtk_entry_get_placeholder_text(v.Native())
	if c == nil {
		return "", nilPtrErr
	}
	return C.GoString((*C.char)(c)), nil
}

func (v *Entry) SetOverwriteMode(overwrite bool) {
	C.gtk_entry_set_overwrite_mode(v.Native(), gbool(overwrite))
}

func (v *Entry) GetOverwriteMode() bool {
	c := C.gtk_entry_get_overwrite_mode(v.Native())
	return gobool(c)
}

// TODO(jrick) Pangolayout
func (v *Entry) GetLayout() {
}

func (v *Entry) GetLayoutOffsets() (x, y int) {
	var gx, gy C.gint
	C.gtk_entry_get_layout_offsets(v.Native(), &gx, &gy)
	return int(gx), int(gy)
}

func (v *Entry) LayoutIndexToTextIndex(layoutIndex int) int {
	c := C.gtk_entry_layout_index_to_text_index(v.Native(),
		C.gint(layoutIndex))
	return int(c)
}

func (v *Entry) TextIndexToLayoutIndex(textIndex int) int {
	c := C.gtk_entry_text_index_to_layout_index(v.Native(),
		C.gint(textIndex))
	return int(c)
}

// TODO(jrick) PandoAttrList
func (v *Entry) SetAttributes() {
}

// TODO(jrick) PandoAttrList
func (v *Entry) GetAttributes() {
}

func (v *Entry) GetMaxLength() int {
	c := C.gtk_entry_get_max_length(v.Native())
	return int(c)
}

func (v *Entry) GetVisibility() bool {
	c := C.gtk_entry_get_visibility(v.Native())
	return gobool(c)
}

func (v *Entry) SetCompletion(completion *EntryCompletion) {
	C.gtk_entry_set_completion(v.Native(), completion.Native())
}

func (v *Entry) GetCompletion() (*EntryCompletion, error) {
	c := C.gtk_entry_get_completion(v.Native())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	e := &EntryCompletion{obj}
	obj.Ref()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return e, nil
}

func (v *Entry) SetCursorHAdjustment(adjustment *Adjustment) {
	C.gtk_entry_set_cursor_hadjustment(v.Native(), adjustment.Native())
}

func (v *Entry) GetCursorHAdjustment() (*Adjustment, error) {
	c := C.gtk_entry_get_cursor_hadjustment(v.Native())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	a := &Adjustment{glib.InitiallyUnowned{obj}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return a, nil
}

func (v *Entry) SetProgressFraction(fraction float64) {
	C.gtk_entry_set_progress_fraction(v.Native(), C.gdouble(fraction))
}

func (v *Entry) GetProgressFraction() float64 {
	c := C.gtk_entry_get_progress_fraction(v.Native())
	return float64(c)
}

func (v *Entry) SetProgressPulseStep(fraction float64) {
	C.gtk_entry_set_progress_pulse_step(v.Native(), C.gdouble(fraction))
}

func (v *Entry) GetProgressPulseStep() float64 {
	c := C.gtk_entry_get_progress_pulse_step(v.Native())
	return float64(c)
}

func (v *Entry) ProgressPulse() {
	C.gtk_entry_progress_pulse(v.Native())
}

// TODO(jrick) GdkEventKey
func (v *Entry) IMContextFilterKeypress() {
}

func (v *Entry) ResetIMContext() {
	C.gtk_entry_reset_im_context(v.Native())
}

// TODO(jrick) GdkPixbuf
func (v *Entry) SetIconFromPixbuf() {
}

func (v *Entry) SetIconFromStock(iconPos EntryIconPosition, stockID string) {
	cstr := C.CString(stockID)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_entry_set_icon_from_stock(v.Native(),
		C.GtkEntryIconPosition(iconPos), (*C.gchar)(cstr))
}

func (v *Entry) SetIconFromIconName(iconPos EntryIconPosition, name string) {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_entry_set_icon_from_icon_name(v.Native(),
		C.GtkEntryIconPosition(iconPos), (*C.gchar)(cstr))
}

// TODO(jrick) GIcon
func (v *Entry) SetIconFromGIcon() {
}

func (v *Entry) GetIconStorageType(iconPos EntryIconPosition) ImageType {
	c := C.gtk_entry_get_icon_storage_type(v.Native(),
		C.GtkEntryIconPosition(iconPos))
	return ImageType(c)
}

// TODO(jrick) GdkPixbuf
func (v *Entry) GetIconPixbuf() {
}

func (v *Entry) GetIconStock(iconPos EntryIconPosition) (string, error) {
	c := C.gtk_entry_get_icon_stock(v.Native(),
		C.GtkEntryIconPosition(iconPos))
	if c == nil {
		return "", nilPtrErr
	}
	return C.GoString((*C.char)(c)), nil
}

func (v *Entry) GetIconName(iconPos EntryIconPosition) (string, error) {
	c := C.gtk_entry_get_icon_name(v.Native(),
		C.GtkEntryIconPosition(iconPos))
	if c == nil {
		return "", nilPtrErr
	}
	return C.GoString((*C.char)(c)), nil
}

// TODO(jrick) GIcon
func (v *Entry) GetIconGIcon() {
}

func (v *Entry) SetIconActivatable(iconPos EntryIconPosition, activatable bool) {
	C.gtk_entry_set_icon_activatable(v.Native(),
		C.GtkEntryIconPosition(iconPos), gbool(activatable))
}

func (v *Entry) GetIconActivatable(iconPos EntryIconPosition) bool {
	c := C.gtk_entry_get_icon_activatable(v.Native(),
		C.GtkEntryIconPosition(iconPos))
	return gobool(c)
}

func (v *Entry) SetIconSensitive(iconPos EntryIconPosition, sensitive bool) {
	C.gtk_entry_set_icon_sensitive(v.Native(),
		C.GtkEntryIconPosition(iconPos), gbool(sensitive))
}

func (v *Entry) GetIconSensitive(iconPos EntryIconPosition) bool {
	c := C.gtk_entry_get_icon_sensitive(v.Native(),
		C.GtkEntryIconPosition(iconPos))
	return gobool(c)
}

func (v *Entry) GetIconAtPos(x, y int) int {
	c := C.gtk_entry_get_icon_at_pos(v.Native(), C.gint(x), C.gint(y))
	return int(c)
}

func (v *Entry) SetIconTooltipText(iconPos EntryIconPosition, tooltip string) {
	cstr := C.CString(tooltip)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_entry_set_icon_tooltip_text(v.Native(),
		C.GtkEntryIconPosition(iconPos), (*C.gchar)(cstr))
}

func (v *Entry) GetIconTooltipText(iconPos EntryIconPosition) (string, error) {
	c := C.gtk_entry_get_icon_tooltip_text(v.Native(),
		C.GtkEntryIconPosition(iconPos))
	if c == nil {
		return "", nilPtrErr
	}
	return C.GoString((*C.char)(c)), nil
}

func (v *Entry) SetIconTooltipMarkup(iconPos EntryIconPosition, tooltip string) {
	cstr := C.CString(tooltip)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_entry_set_icon_tooltip_markup(v.Native(),
		C.GtkEntryIconPosition(iconPos), (*C.gchar)(cstr))
}

func (v *Entry) GetIconTooltipMarkup(iconPos EntryIconPosition) (string, error) {
	c := C.gtk_entry_get_icon_tooltip_markup(v.Native(),
		C.GtkEntryIconPosition(iconPos))
	if c == nil {
		return "", nilPtrErr
	}
	return C.GoString((*C.char)(c)), nil
}

// TODO(jrick) GdkDragAction
func (v *Entry) SetIconDragSource() {
}

func (v *Entry) GetCurrentIconDragSource() int {
	c := C.gtk_entry_get_current_icon_drag_source(v.Native())
	return int(c)
}

// TODO(jrick) GdkRectangle
func (v *Entry) GetIconArea() {
}

func (v *Entry) SetInputPurpose(purpose InputPurpose) {
	C.gtk_entry_set_input_purpose(v.Native(), C.GtkInputPurpose(purpose))
}

func (v *Entry) GetInputPurpose() InputPurpose {
	c := C.gtk_entry_get_input_purpose(v.Native())
	return InputPurpose(c)
}

func (v *Entry) SetInputHints(hints InputHints) {
	C.gtk_entry_set_input_hints(v.Native(), C.GtkInputHints(hints))
}

func (v *Entry) GetInputHints() InputHints {
	c := C.gtk_entry_get_input_hints(v.Native())
	return InputHints(c)
}

/*
 * GtkEntryBuffer
 */

type EntryBuffer struct {
	*glib.Object
}

func (v *EntryBuffer) Native() *C.GtkEntryBuffer {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkEntryBuffer(p)
}

func EntryBufferNew(initialChars string, nInitialChars int) (*EntryBuffer, error) {
	cstr := C.CString(initialChars)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_entry_buffer_new((*C.gchar)(cstr), C.gint(nInitialChars))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	e := &EntryBuffer{obj}
	obj.Ref()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return e, nil
}

func (v *EntryBuffer) GetText() (string, error) {
	c := C.gtk_entry_buffer_get_text(v.Native())
	if c == nil {
		return "", nilPtrErr
	}
	return C.GoString((*C.char)(c)), nil
}

func (v *EntryBuffer) SetText(text string) {
	cstr := C.CString(text)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_entry_buffer_set_text(v.Native(), (*C.gchar)(cstr),
		C.gint(len(text)))
}

func (v *EntryBuffer) GetBytes() uint {
	c := C.gtk_entry_buffer_get_bytes(v.Native())
	return uint(c)
}

func (v *EntryBuffer) GetLength() uint {
	c := C.gtk_entry_buffer_get_length(v.Native())
	return uint(c)
}

func (v *EntryBuffer) GetMaxLength() int {
	c := C.gtk_entry_buffer_get_max_length(v.Native())
	return int(c)
}

func (v *EntryBuffer) SetMaxLength(maxLength int) {
	C.gtk_entry_buffer_set_max_length(v.Native(), C.gint(maxLength))
}

func (v *EntryBuffer) InsertText(position uint, text string) uint {
	cstr := C.CString(text)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_entry_buffer_insert_text(v.Native(), C.guint(position),
		(*C.gchar)(cstr), C.gint(len(text)))
	return uint(c)
}

func (v *EntryBuffer) DeleteText(position uint, nChars int) uint {
	c := C.gtk_entry_buffer_delete_text(v.Native(), C.guint(position),
		C.gint(nChars))
	return uint(c)
}

func (v *EntryBuffer) EmitDeletedText(pos, nChars uint) {
	C.gtk_entry_buffer_emit_deleted_text(v.Native(), C.guint(pos),
		C.guint(nChars))
}

func (v *EntryBuffer) EmitInsertedText(pos uint, text string) {
	cstr := C.CString(text)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_entry_buffer_emit_inserted_text(v.Native(), C.guint(pos),
		(*C.gchar)(cstr), C.guint(len(text)))
}

/*
 * GtkEntryCompletion
 */

type EntryCompletion struct {
	*glib.Object
}

func (v *EntryCompletion) Native() *C.GtkEntryCompletion {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkEntryCompletion(p)
}

/*
 * GtkGrid
 */

type Grid struct {
	Container

	// Interfaces
	Orientable
}

func (v *Grid) Native() *C.GtkGrid {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkGrid(p)
}

func (v *Grid) toOrientable() *C.GtkOrientable {
	if v == nil {
		return nil
	}
	return C.toGtkOrientable(unsafe.Pointer(v.GObject))
}

func GridNew() (*Grid, error) {
	c := C.gtk_grid_new()
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	o := Orientable{obj}
	g := &Grid{Container{Widget{glib.InitiallyUnowned{obj}}}, o}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return g, nil
}

func (v *Grid) Attach(child IWidget, left, top, width, height int) {
	C.gtk_grid_attach(v.Native(), child.toWidget(), C.gint(left),
		C.gint(top), C.gint(width), C.gint(height))
}

func (v *Grid) AttachNextTo(child, sibling IWidget, side PositionType, width, height int) {
	C.gtk_grid_attach_next_to(v.Native(), child.toWidget(),
		sibling.toWidget(), C.GtkPositionType(side), C.gint(width),
		C.gint(height))
}

func (v *Grid) GetChildAt(left, top int) (*Widget, error) {
	c := C.gtk_grid_get_child_at(v.Native(), C.gint(left), C.gint(top))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	w := &Widget{glib.InitiallyUnowned{obj}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return w, nil
}

func (v *Grid) InsertRow(position int) {
	C.gtk_grid_insert_row(v.Native(), C.gint(position))
}

func (v *Grid) InsertColumn(position int) {
	C.gtk_grid_insert_column(v.Native(), C.gint(position))
}

func (v *Grid) InsertNextTo(sibling IWidget, side PositionType) {
	C.gtk_grid_insert_next_to(v.Native(), sibling.toWidget(),
		C.GtkPositionType(side))
}

func (v *Grid) SetRowHomogeneous(homogeneous bool) {
	C.gtk_grid_set_row_homogeneous(v.Native(), gbool(homogeneous))
}

func (v *Grid) GetRowHomogeneous() bool {
	c := C.gtk_grid_get_row_homogeneous(v.Native())
	return gobool(c)
}

func (v *Grid) SetRowSpacing(spacing uint) {
	C.gtk_grid_set_row_spacing(v.Native(), C.guint(spacing))
}

func (v *Grid) GetRowSpacing() uint {
	c := C.gtk_grid_get_row_spacing(v.Native())
	return uint(c)
}

func (v *Grid) SetColumnHomogeneous(homogeneous bool) {
	C.gtk_grid_set_column_homogeneous(v.Native(), gbool(homogeneous))
}

func (v *Grid) GetColumnHomogeneous() bool {
	c := C.gtk_grid_get_column_homogeneous(v.Native())
	return gobool(c)
}

func (v *Grid) SetColumnSpacing(spacing uint) {
	C.gtk_grid_set_column_spacing(v.Native(), C.guint(spacing))
}

func (v *Grid) GetColumnSpacing() uint {
	c := C.gtk_grid_get_column_spacing(v.Native())
	return uint(c)
}

/*
 * GtkImage
 */

type Image struct {
	Misc
}

func (v *Image) Native() *C.GtkImage {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkImage(p)
}

func ImageNew() (*Image, error) {
	c := C.gtk_image_new()
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	i := &Image{Misc{Widget{glib.InitiallyUnowned{obj}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return i, nil
}

func ImageNewFromFile(filename string) (*Image, error) {
	cstr := C.CString(filename)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_image_new_from_file((*C.gchar)(cstr))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	i := &Image{Misc{Widget{glib.InitiallyUnowned{obj}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return i, nil
}

func ImageNewFromResource(resourcePath string) (*Image, error) {
	cstr := C.CString(resourcePath)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_image_new_from_resource((*C.gchar)(cstr))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	i := &Image{Misc{Widget{glib.InitiallyUnowned{obj}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return i, nil
}

// TODO(jrick) GdkPixbuf
func ImageNewFromPixbuf() {
}

func ImageNewFromStock(stock Stock, size IconSize) (*Image, error) {
	cstr := C.CString(string(stock))
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_image_new_from_stock((*C.gchar)(cstr), C.GtkIconSize(size))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	i := &Image{Misc{Widget{glib.InitiallyUnowned{obj}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return i, nil
}

// TODO(jrick) GtkIconSet
func ImageNewFromIconSet() {
}

// TODO(jrick) GdkPixbufAnimation
func ImageNewFromAnimation() {
}

func ImageNewFromIconName(iconName string, size IconSize) (*Image, error) {
	cstr := C.CString(iconName)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_image_new_from_icon_name((*C.gchar)(cstr),
		C.GtkIconSize(size))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	i := &Image{Misc{Widget{glib.InitiallyUnowned{obj}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return i, nil
}

// TODO(jrick) GIcon
func ImageNewFromGIcon() {
}

func (v *Image) Clear() {
	C.gtk_image_clear(v.Native())
}

func (v *Image) SetFromFile(filename string) {
	cstr := C.CString(filename)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_image_set_from_file(v.Native(), (*C.gchar)(cstr))
}

func (v *Image) SetFromResource(resourcePath string) {
	cstr := C.CString(resourcePath)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_image_set_from_resource(v.Native(), (*C.gchar)(cstr))
}

// TODO(jrick) GdkPixbuf
func (v *Image) SetFromPixbuf() {
}

func (v *Image) SetFromStock(stock Stock, size IconSize) {
	cstr := C.CString(string(stock))
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_image_set_from_stock(v.Native(), (*C.gchar)(cstr),
		C.GtkIconSize(size))
}

// TODO(jrick) GtkIconSet
func (v *Image) SetFromIconSet() {
}

// TODO(jrick) GdkPixbufAnimation
func (v *Image) SetFromAnimation() {
}

func (v *Image) SetFromIconName(iconName string, size IconSize) {
	cstr := C.CString(iconName)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_image_set_from_icon_name(v.Native(), (*C.gchar)(cstr),
		C.GtkIconSize(size))
}

// TODO(jrick) GIcon
func (v *Image) SetFromGIcon() {
}

func (v *Image) SetPixelSize(pixelSize int) {
	C.gtk_image_set_pixel_size(v.Native(), C.gint(pixelSize))
}

func (v *Image) GetStorageType() ImageType {
	c := C.gtk_image_get_storage_type(v.Native())
	return ImageType(c)
}

// TODO(jrick) GdkPixbuf
func (v *Image) GetPixbuf() {
}

// TODO(jrick) GtkIconSet
func (v *Image) GetIconSet() {
}

// TODO(jrick) GdkPixbufAnimation
func (v *Image) GetAnimation() {
}

func (v *Image) GetIconName() (string, IconSize) {
	var iconName *C.gchar
	var size C.GtkIconSize
	C.gtk_image_get_icon_name(v.Native(), &iconName, &size)
	return C.GoString((*C.char)(iconName)), IconSize(size)
}

// TODO(jrick) GIcon
func (v *Image) GetGIcon() {
}

func (v *Image) GetPixelSize() int {
	c := C.gtk_image_get_pixel_size(v.Native())
	return int(c)
}

/*
 * GtkLabel
 */

type Label struct {
	Misc
}

func (v *Label) Native() *C.GtkLabel {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkLabel(p)
}

func LabelNew(str string) (*Label, error) {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_label_new((*C.gchar)(cstr))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	l := &Label{Misc{Widget{glib.InitiallyUnowned{obj}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return l, nil
}

func (v *Label) SetText(str string) {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_label_set_text(v.Native(), (*C.gchar)(cstr))
}

func (v *Label) SetMarkup(str string) {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_label_set_markup(v.Native(), (*C.gchar)(cstr))
}

func (v *Label) SetMarkupWithMnemonic(str string) {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_label_set_markup_with_mnemonic(v.Native(), (*C.gchar)(cstr))
}

func (v *Label) SetPattern(patern string) {
	cstr := C.CString(patern)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_label_set_pattern(v.Native(), (*C.gchar)(cstr))
}

func (v *Label) SetWidthChars(nChars int) {
	C.gtk_label_set_width_chars(v.Native(), C.gint(nChars))
}

func (v *Label) SetMaxWidthChars(nChars int) {
	C.gtk_label_set_max_width_chars(v.Native(), C.gint(nChars))
}

func (v *Label) SetLineWrap(wrap bool) {
	C.gtk_label_set_line_wrap(v.Native(), gbool(wrap))
}

func (v *Label) GetSelectable() bool {
	c := C.gtk_label_get_selectable(v.Native())
	return gobool(c)
}

func (v *Label) GetText() (string, error) {
	c := C.gtk_label_get_text(v.Native())
	if c == nil {
		return "", nilPtrErr
	}
	return C.GoString((*C.char)(c)), nil
}

func LabelNewWithMnemonic(str string) (*Label, error) {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_label_new_with_mnemonic((*C.gchar)(cstr))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	l := &Label{Misc{Widget{glib.InitiallyUnowned{obj}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return l, nil
}

func (v *Label) SetSelectable(setting bool) {
	C.gtk_label_set_selectable(v.Native(), gbool(setting))
}

func (v *Label) SetLabel(str string) {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_label_set_label(v.Native(), (*C.gchar)(cstr))
}

/*
 * GtkListStore
 */

type ListStore struct {
	*glib.Object

	// Interfaces
	TreeModel
}

func (v *ListStore) Native() *C.GtkListStore {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkListStore(p)
}

func (v *ListStore) toTreeModel() *C.GtkTreeModel {
	if v == nil {
		return nil
	}
	return C.toGtkTreeModel(unsafe.Pointer(v.GObject))
}

func ListStoreNew(types ...glib.Type) (*ListStore, error) {
	gtypes := C.alloc_types(C.int(len(types)))
	for n, val := range types {
		C.set_type(gtypes, C.int(n), C.GType(val))
	}
	defer C.g_free(C.gpointer(gtypes))
	c := C.gtk_list_store_newv(C.gint(len(types)), gtypes)
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	tm := TreeModel{obj}
	ls := &ListStore{obj, tm}
	obj.Ref()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return ls, nil
}

// TODO(jrick)
func (v *ListStore) SetColumnTypes(types ...glib.Type) {
}

func (v *ListStore) Set(iter *TreeIter, columns []int, values []interface{}) error {
	if len(columns) != len(values) {
		return errors.New("columns and values lengths do not match")
	}
	for i, val := range values {
		if gv, err := glib.GValue(val); err != nil {
			return err
		} else {
			C.gtk_list_store_set_value(v.Native(), iter.Native(),
				C.gint(columns[i]),
				(*C.GValue)(unsafe.Pointer(gv.Native())))
		}
	}
	return nil
}

// TODO(jrick)
func (v *ListStore) InsertWithValues(iter *TreeIter, position int, columns []int, values []glib.Value) {
	/*
		var ccolumns *C.gint
		var cvalues *C.GValue

		C.gtk_list_store_insert_with_values(v.native(), iter.Native(),
			C.gint(position), columns, values, C.gint(len(values)))
	*/
}

func (v *ListStore) Prepend(iter *TreeIter) {
	C.gtk_list_store_prepend(v.Native(), iter.Native())
}

func (v *ListStore) Append(iter *TreeIter) {
	C.gtk_list_store_append(v.Native(), iter.Native())
}

func (v *ListStore) Clear() {
	C.gtk_list_store_clear(v.Native())
}

func (v *ListStore) IterIsValid(iter *TreeIter) bool {
	c := C.gtk_list_store_iter_is_valid(v.Native(), iter.Native())
	return gobool(c)
}

// TODO(jrick)
func (v *ListStore) Reorder(newOrder []int) {
}

func (v *ListStore) Swap(a, b *TreeIter) {
	C.gtk_list_store_swap(v.Native(), a.Native(), b.Native())
}

func (v *ListStore) MoveBefore(iter, position *TreeIter) {
	C.gtk_list_store_move_before(v.Native(), iter.Native(),
		position.Native())
}

func (v *ListStore) MoveAfter(iter, position *TreeIter) {
	C.gtk_list_store_move_after(v.Native(), iter.Native(),
		position.Native())
}

/*
 * GtkMenu
 */

type Menu struct {
	MenuShell
}

func (v *Menu) Native() *C.GtkMenu {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkMenu(p)
}

func MenuNew() (*Menu, error) {
	c := C.gtk_menu_new()
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	m := &Menu{MenuShell{Container{Widget{glib.InitiallyUnowned{obj}}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return m, nil
}

/*
 * GtkMenuBar
 */

type MenuBar struct {
	MenuShell
}

func (v *MenuBar) Native() *C.GtkMenuBar {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkMenuBar(p)
}

func MenuBarNew() (*MenuBar, error) {
	c := C.gtk_menu_bar_new()
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	m := &MenuBar{MenuShell{Container{Widget{glib.InitiallyUnowned{obj}}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return m, nil
}

/*
 * GtkMenuItem
 */

type MenuItem struct {
	Bin
}

func (v *MenuItem) Native() *C.GtkMenuItem {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkMenuItem(p)
}

func MenuItemNew() (*MenuItem, error) {
	c := C.gtk_menu_item_new()
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	m := &MenuItem{Bin{Container{Widget{glib.InitiallyUnowned{obj}}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return m, nil
}

func MenuItemNewWithLabel(label string) (*MenuItem, error) {
	cstr := C.CString(label)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_menu_item_new_with_label((*C.gchar)(cstr))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	m := &MenuItem{Bin{Container{Widget{glib.InitiallyUnowned{obj}}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return m, nil
}

func MenuItemNewWithMnemonic(label string) (*MenuItem, error) {
	cstr := C.CString(label)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_menu_item_new_with_mnemonic((*C.gchar)(cstr))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	m := &MenuItem{Bin{Container{Widget{glib.InitiallyUnowned{obj}}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return m, nil
}

func (v *MenuItem) SetSubmenu(submenu IWidget) {
	C.gtk_menu_item_set_submenu(v.Native(), submenu.toWidget())
}

/*
 * GtkMenuShell
 */

type MenuShell struct {
	Container
}

func (v *MenuShell) Native() *C.GtkMenuShell {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkMenuShell(p)
}

func (v *MenuShell) Append(child IWidget) {
	C.gtk_menu_shell_append(v.Native(), child.toWidget())
}

/*
 * GtkMessageDialog
 */

type MessageDialog struct {
	Dialog
}

func (v *MessageDialog) Native() *C.GtkMessageDialog {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkMessageDialog(p)
}

func MessageDialogNew(parent IWindow, flags DialogFlags, mType MessageType, buttons ButtonsType, format string, a ...interface{}) *MessageDialog {
	s := fmt.Sprintf(format, a...)
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr))
	c := C._gtk_message_dialog_new(parent.toWindow(),
		C.GtkDialogFlags(flags), C.GtkMessageType(mType),
		C.GtkButtonsType(buttons), cstr)
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	m := &MessageDialog{Dialog{Window{Bin{Container{Widget{
		glib.InitiallyUnowned{obj}}}}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return m
}

/*
 * GtkMisc
 */

type Misc struct {
	Widget
}

func (v *Misc) Native() *C.GtkMisc {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkMisc(p)
}

/*
 * GtkNotebook
 */

type Notebook struct {
	Container
}

func (v *Notebook) Native() *C.GtkNotebook {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkNotebook(p)
}

func NotebookNew() (*Notebook, error) {
	c := C.gtk_notebook_new()
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	n := &Notebook{Container{Widget{glib.InitiallyUnowned{obj}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return n, nil
}

func (v *Notebook) AppendPage(child IWidget, tabLabel IWidget) int {
	c := C.gtk_notebook_append_page(v.Native(), child.toWidget(),
		tabLabel.toWidget())
	return int(c)
}

func (v *Notebook) AppendPageMenu(child IWidget, tabLabel IWidget, menuLabel IWidget) int {
	c := C.gtk_notebook_append_page_menu(v.Native(), child.toWidget(),
		tabLabel.toWidget(), menuLabel.toWidget())
	return int(c)
}

func (v *Notebook) PrependPage(child IWidget, tabLabel IWidget) int {
	c := C.gtk_notebook_prepend_page(v.Native(), child.toWidget(),
		tabLabel.toWidget())
	return int(c)
}

func (v *Notebook) PrependPageMenu(child IWidget, tabLabel IWidget, menuLabel IWidget) int {
	c := C.gtk_notebook_prepend_page_menu(v.Native(), child.toWidget(),
		tabLabel.toWidget(), menuLabel.toWidget())
	return int(c)
}

func (v *Notebook) InsertPage(child IWidget, tabLabel IWidget, position int) int {
	c := C.gtk_notebook_insert_page(v.Native(), child.toWidget(),
		tabLabel.toWidget(), C.gint(position))
	return int(c)
}

func (v *Notebook) InsertPageMenu(child IWidget, tabLabel IWidget, menuLabel IWidget, position int) int {
	c := C.gtk_notebook_insert_page_menu(v.Native(), child.toWidget(),
		tabLabel.toWidget(), menuLabel.toWidget(), C.gint(position))
	return int(c)
}

func (v *Notebook) RemovePage(pageNum int) {
	C.gtk_notebook_remove_page(v.Native(), C.gint(pageNum))
}

func (v *Notebook) PageNum(child IWidget) int {
	c := C.gtk_notebook_page_num(v.Native(), child.toWidget())
	return int(c)
}

func (v *Notebook) NextPage() {
	C.gtk_notebook_next_page(v.Native())
}

func (v *Notebook) PrevPage() {
	C.gtk_notebook_prev_page(v.Native())
}

func (v *Notebook) ReorderChild(child IWidget, position int) {
	C.gtk_notebook_reorder_child(v.Native(), child.toWidget(),
		C.gint(position))
}

func (v *Notebook) SetTabPos(pos PositionType) {
	C.gtk_notebook_set_tab_pos(v.Native(), C.GtkPositionType(pos))
}

func (v *Notebook) SetShowTabs(showTabs bool) {
	C.gtk_notebook_set_show_tabs(v.Native(), gbool(showTabs))
}

func (v *Notebook) SetShowBorder(showBorder bool) {
	C.gtk_notebook_set_show_border(v.Native(), gbool(showBorder))
}

func (v *Notebook) SetScrollable(scrollable bool) {
	C.gtk_notebook_set_scrollable(v.Native(), gbool(scrollable))
}

func (v *Notebook) PopupEnable() {
	C.gtk_notebook_popup_enable(v.Native())
}

func (v *Notebook) PopupDisable() {
	C.gtk_notebook_popup_disable(v.Native())
}

func (v *Notebook) GetCurrentPage() int {
	c := C.gtk_notebook_get_current_page(v.Native())
	return int(c)
}

func (v *Notebook) GetMenuLabel(child IWidget) (*Widget, error) {
	c := C.gtk_notebook_get_menu_label(v.Native(), child.toWidget())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	w := &Widget{glib.InitiallyUnowned{obj}}
	w.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return w, nil
}

func (v *Notebook) GetNthPage(pageNum int) (*Widget, error) {
	c := C.gtk_notebook_get_nth_page(v.Native(), C.gint(pageNum))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	w := &Widget{glib.InitiallyUnowned{obj}}
	w.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return w, nil
}

func (v *Notebook) GetNPages() int {
	c := C.gtk_notebook_get_n_pages(v.Native())
	return int(c)
}

func (v *Notebook) GetTabLabel(child IWidget) (*Widget, error) {
	c := C.gtk_notebook_get_tab_label(v.Native(), child.toWidget())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	w := &Widget{glib.InitiallyUnowned{obj}}
	w.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return w, nil
}

func (v *Notebook) SetMenuLabel(child, menuLabel IWidget) {
	C.gtk_notebook_set_menu_label(v.Native(), child.toWidget(),
		menuLabel.toWidget())
}

func (v *Notebook) SetMenuLabelText(child IWidget, menuText string) {
	cstr := C.CString(menuText)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_notebook_set_menu_label_text(v.Native(), child.toWidget(),
		(*C.gchar)(cstr))
}

func (v *Notebook) SetTabLabel(child, tabLabel IWidget) {
	C.gtk_notebook_set_tab_label(v.Native(), child.toWidget(),
		tabLabel.toWidget())
}

func (v *Notebook) SetTabLabelText(child IWidget, tabText string) {
	cstr := C.CString(tabText)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_notebook_set_tab_label_text(v.Native(), child.toWidget(),
		(*C.gchar)(cstr))
}

func (v *Notebook) SetTabReorderable(child IWidget, reorderable bool) {
	C.gtk_notebook_set_tab_reorderable(v.Native(), child.toWidget(),
		gbool(reorderable))
}

func (v *Notebook) SetTabDetachable(child IWidget, detachable bool) {
	C.gtk_notebook_set_tab_detachable(v.Native(), child.toWidget(),
		gbool(detachable))
}

func (v *Notebook) GetMenuLabelText(child IWidget) (string, error) {
	c := C.gtk_notebook_get_menu_label_text(v.Native(), child.toWidget())
	if c == nil {
		return "", errors.New("No menu label for widget")
	}
	return C.GoString((*C.char)(c)), nil
}

func (v *Notebook) GetScrollable() bool {
	c := C.gtk_notebook_get_scrollable(v.Native())
	return gobool(c)
}

func (v *Notebook) GetShowBorder() bool {
	c := C.gtk_notebook_get_show_border(v.Native())
	return gobool(c)
}

func (v *Notebook) GetShowTabs() bool {
	c := C.gtk_notebook_get_show_tabs(v.Native())
	return gobool(c)
}

func (v *Notebook) GetTabLabelText(child IWidget) (string, error) {
	c := C.gtk_notebook_get_tab_label_text(v.Native(), child.toWidget())
	if c == nil {
		return "", errors.New("No tab label for widget")
	}
	return C.GoString((*C.char)(c)), nil
}

func (v *Notebook) GetTabPos() PositionType {
	c := C.gtk_notebook_get_tab_pos(v.Native())
	return PositionType(c)
}

func (v *Notebook) GetTabReorderable(child IWidget) bool {
	c := C.gtk_notebook_get_tab_reorderable(v.Native(), child.toWidget())
	return gobool(c)
}

func (v *Notebook) GetTabDetachable(child IWidget) bool {
	c := C.gtk_notebook_get_tab_detachable(v.Native(), child.toWidget())
	return gobool(c)
}

func (v *Notebook) SetCurrentPage(pageNum int) {
	C.gtk_notebook_set_current_page(v.Native(), C.gint(pageNum))
}

func (v *Notebook) SetGroupName(groupName string) {
	cstr := C.CString(groupName)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_notebook_set_group_name(v.Native(), (*C.gchar)(cstr))
}

func (v *Notebook) GetGroupName() (string, error) {
	c := C.gtk_notebook_get_group_name(v.Native())
	if c == nil {
		return "", errors.New("No group name")
	}
	return C.GoString((*C.char)(c)), nil
}

func (v *Notebook) SetActionWidget(widget IWidget, packType PackType) {
	C.gtk_notebook_set_action_widget(v.Native(), widget.toWidget(),
		C.GtkPackType(packType))
}

func (v *Notebook) GetActionWidget(packType PackType) (*Widget, error) {
	c := C.gtk_notebook_get_action_widget(v.Native(),
		C.GtkPackType(packType))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	w := &Widget{glib.InitiallyUnowned{obj}}
	w.RefSink()
	runtime.SetFinalizer(w, (*glib.Object).Unref)
	return w, nil
}

/*
 * GtkOrientable
 */

type Orientable struct {
	*glib.Object
}

type IOrientable interface {
	toOrientable() *C.GtkOrientable
}

func (v *Orientable) Native() *C.GtkOrientable {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkOrientable(p)
}

func (v *Orientable) GetOrientation() Orientation {
	c := C.gtk_orientable_get_orientation(v.Native())
	return Orientation(c)
}

func (v *Orientable) SetOrientation(orientation Orientation) {
	C.gtk_orientable_set_orientation(v.Native(),
		C.GtkOrientation(orientation))
}

/*
 * GtkProgressBar
 */

type ProgressBar struct {
	Widget
}

func (v *ProgressBar) Native() *C.GtkProgressBar {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkProgressBar(p)
}

func ProgressBarNew() (*ProgressBar, error) {
	c := C.gtk_progress_bar_new()
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	p := &ProgressBar{Widget{glib.InitiallyUnowned{obj}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return p, nil
}

func (v *ProgressBar) SetFraction(fraction float64) {
	C.gtk_progress_bar_set_fraction(v.Native(), C.gdouble(fraction))
}

func (v *ProgressBar) GetFraction() float64 {
	c := C.gtk_progress_bar_get_fraction(v.Native())
	return float64(c)
}

func (v *ProgressBar) SetText(text string) {
	cstr := C.CString(text)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_progress_bar_set_text(v.Native(), (*C.gchar)(cstr))
}

/*
 * GtkScrolledWindow
 */

type ScrolledWindow struct {
	Bin
}

func (v *ScrolledWindow) Native() *C.GtkScrolledWindow {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkScrolledWindow(p)
}

func ScrolledWindowNew(hadjustment, vadjustment *Adjustment) (*ScrolledWindow, error) {
	c := C.gtk_scrolled_window_new(hadjustment.Native(),
		vadjustment.Native())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	s := &ScrolledWindow{Bin{Container{Widget{glib.InitiallyUnowned{obj}}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return s, nil
}

func (v *ScrolledWindow) SetPolicy(hScrollbarPolicy, vScrollbarPolicy PolicyType) {
	C.gtk_scrolled_window_set_policy(v.Native(),
		C.GtkPolicyType(hScrollbarPolicy),
		C.GtkPolicyType(vScrollbarPolicy))
}

// others...

/*
 * GtkSpinButton
 */

type SpinButton struct {
	Entry
}

func (v *SpinButton) Native() *C.GtkSpinButton {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkSpinButton(p)
}

func (v *SpinButton) Configure(adjustment *Adjustment, climbRate float64, digits uint) {
	C.gtk_spin_button_configure(v.Native(), adjustment.Native(),
		C.gdouble(climbRate), C.guint(digits))
}

func SpinButtonNew(adjustment *Adjustment, climbRate float64, digits uint) (*SpinButton, error) {
	c := C.gtk_spin_button_new(adjustment.Native(),
		C.gdouble(climbRate), C.guint(digits))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	s := &SpinButton{Entry{Widget{glib.InitiallyUnowned{obj}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return s, nil
}

func SpinButtonNewWithRange(min, max, step float64) (*SpinButton, error) {
	c := C.gtk_spin_button_new_with_range(C.gdouble(min), C.gdouble(max),
		C.gdouble(step))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	s := &SpinButton{Entry{Widget{glib.InitiallyUnowned{obj}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return s, nil
}

func (v *SpinButton) GetValueAsInt() int {
	c := C.gtk_spin_button_get_value_as_int(v.Native())
	return int(c)
}

func (v *SpinButton) GetValue() float64 {
	c := C.gtk_spin_button_get_value(v.Native())
	return float64(c)
}

/*
 * GtkStatusbar
 */

type Statusbar struct {
	Box
}

func (v *Statusbar) Native() *C.GtkStatusbar {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkStatusbar(p)
}

func StatusbarNew() (*Statusbar, error) {
	c := C.gtk_statusbar_new()
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	s := &Statusbar{Box{Container{Widget{glib.InitiallyUnowned{obj}}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return s, nil
}

func (v *Statusbar) GetContextId(contextDescription string) uint {
	cstr := C.CString(contextDescription)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_statusbar_get_context_id(v.Native(), (*C.gchar)(cstr))
	return uint(c)
}

func (v *Statusbar) Push(contextID uint, text string) uint {
	cstr := C.CString(text)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_statusbar_push(v.Native(), C.guint(contextID),
		(*C.gchar)(cstr))
	return uint(c)
}

func (v *Statusbar) Pop(contextID uint) {
	C.gtk_statusbar_pop(v.Native(), C.guint(contextID))
}

// others...

/*
 * GtkTreeIter
 */

type TreeIter struct {
	GtkTreeIter C.GtkTreeIter
}

func (v *TreeIter) Native() *C.GtkTreeIter {
	return &v.GtkTreeIter
}

func (v *TreeIter) free() {
	C.gtk_tree_iter_free(v.Native())
}

func (v *TreeIter) Copy() (*TreeIter, error) {
	c := C.gtk_tree_iter_copy(v.Native())
	if c == nil {
		return nil, nilPtrErr
	}
	t := &TreeIter{*c}
	runtime.SetFinalizer(t, (*TreeIter).free)
	return t, nil
}

/*
 * GtkTreeModel
 */

type TreeModel struct {
	*glib.Object
}

type ITreeModel interface {
	toTreeModel() *C.GtkTreeModel
}

func (v *TreeModel) Native() *C.GtkTreeModel {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkTreeModel(p)
}

func (v *TreeModel) toTreeModel() *C.GtkTreeModel {
	if v == nil {
		return nil
	}
	return v.Native()
}

func (v *TreeModel) GetFlags() TreeModelFlags {
	c := C.gtk_tree_model_get_flags(v.Native())
	return TreeModelFlags(c)
}

func (v *TreeModel) GetNColumns() int {
	c := C.gtk_tree_model_get_n_columns(v.Native())
	return int(c)
}

func (v *TreeModel) GetColumnType(index int) glib.Type {
	c := C.gtk_tree_model_get_column_type(v.Native(), C.gint(index))
	return glib.Type(c)
}

func (v *TreeModel) GetIter(path *TreePath) (*TreeIter, error) {
	var iter C.GtkTreeIter
	c := C.gtk_tree_model_get_iter(v.Native(), &iter, path.Native())
	if !gobool(c) {
		return nil, errors.New("Unable to set iterator")
	}
	t := &TreeIter{iter}
	runtime.SetFinalizer(t, (*TreeIter).free)
	return t, nil
}

func (v *TreeModel) GetIterFromString(path string) (*TreeIter, error) {
	var iter C.GtkTreeIter
	cstr := C.CString(path)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_tree_model_get_iter_from_string(v.Native(), &iter,
		(*C.gchar)(cstr))
	if !gobool(c) {
		return nil, errors.New("Unable to set iterator")
	}
	t := &TreeIter{iter}
	runtime.SetFinalizer(t, (*TreeIter).free)
	return t, nil
}

func (v *TreeModel) GetIterFirst() (*TreeIter, error) {
	var iter C.GtkTreeIter
	c := C.gtk_tree_model_get_iter_first(v.Native(), &iter)
	if !gobool(c) {
		return nil, errors.New("Unable to set iterator")
	}
	t := &TreeIter{iter}
	runtime.SetFinalizer(t, (*TreeIter).free)
	return t, nil
}

func (v *TreeModel) GetPath(iter *TreeIter) (*TreePath, error) {
	c := C.gtk_tree_model_get_path(v.Native(), iter.Native())
	if c == nil {
		return nil, nilPtrErr
	}
	p := &TreePath{c}
	runtime.SetFinalizer(p, (*TreePath).free)
	return p, nil
}

func (v *TreeModel) GetValue(iter *TreeIter, column int) (*glib.Value, error) {
	val, err := glib.ValueAlloc()
	if err != nil {
		return nil, err
	}
	C.gtk_tree_model_get_value(
		(*C.GtkTreeModel)(unsafe.Pointer(v.Native())),
		iter.Native(),
		C.gint(column),
		(*C.GValue)(unsafe.Pointer(val.Native())))
	return val, nil
}

/*
 * GtkTreePath
 */

type TreePath struct {
	GtkTreePath *C.GtkTreePath
}

func (v *TreePath) Native() *C.GtkTreePath {
	if v == nil {
		return nil
	}
	return v.GtkTreePath
}

func (v *TreePath) free() {
	C.gtk_tree_path_free(v.Native())
}

/*
 * GtkTreeSelection
 */

type TreeSelection struct {
	*glib.Object
}

func (v *TreeSelection) Native() *C.GtkTreeSelection {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkTreeSelection(p)
}

func (v *TreeSelection) GetSelected(model *ITreeModel, iter *TreeIter) bool {
	var pcmodel **C.GtkTreeModel
	if pcmodel != nil {
		cmodel := (*model).toTreeModel()
		pcmodel = &cmodel
	} else {
		pcmodel = nil
	}
	c := C.gtk_tree_selection_get_selected(v.Native(),
		pcmodel, iter.Native())
	return gobool(c)
}

/*
 * GtkTreeView
 */

type TreeView struct {
	Container
}

func (v *TreeView) Native() *C.GtkTreeView {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkTreeView(p)
}

func TreeViewNew() (*TreeView, error) {
	c := C.gtk_tree_view_new()
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	t := &TreeView{Container{Widget{glib.InitiallyUnowned{obj}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return t, nil
}

func TreeViewNewWithModel(model ITreeModel) (*TreeView, error) {
	c := C.gtk_tree_view_new_with_model(model.toTreeModel())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	t := &TreeView{Container{Widget{glib.InitiallyUnowned{obj}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return t, nil
}

func (v *TreeView) GetModel() (*TreeModel, error) {
	c := C.gtk_tree_view_get_model(v.Native())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	t := &TreeModel{obj}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return t, nil
}

func (v *TreeView) SetModel(model ITreeModel) {
	C.gtk_tree_view_set_model(v.Native(), model.toTreeModel())
}

func (v *TreeView) GetSelection() (*TreeSelection, error) {
	c := C.gtk_tree_view_get_selection(v.Native())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	s := &TreeSelection{obj}
	obj.Ref()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return s, nil
}

func (v *TreeView) AppendColumn(column *TreeViewColumn) int {
	c := C.gtk_tree_view_append_column(v.Native(), column.Native())
	return int(c)
}

/*
 * GtkTreeViewColumn
 */

type TreeViewColumn struct {
	glib.InitiallyUnowned
}

func (v *TreeViewColumn) Native() *C.GtkTreeViewColumn {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkTreeViewColumn(p)
}

func TreeViewColumnNew() (*TreeViewColumn, error) {
	c := C.gtk_tree_view_column_new()
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	t := &TreeViewColumn{glib.InitiallyUnowned{obj}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return t, nil
}

func TreeViewColumnNewWithAttributes(title string, renderer ICellRenderer, attribute string, column int) (*TreeViewColumn, error) {
	t_cstr := C.CString(title)
	defer C.free(unsafe.Pointer(t_cstr))
	a_cstr := C.CString(attribute)
	defer C.free(unsafe.Pointer(a_cstr))
	c := C._gtk_tree_view_column_new_with_attributes_one((*C.gchar)(t_cstr),
		renderer.toCellRenderer(), (*C.gchar)(a_cstr), C.gint(column))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	t := &TreeViewColumn{glib.InitiallyUnowned{obj}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return t, nil
}

func (v *TreeViewColumn) AddAttribute(renderer ICellRenderer, attribute string, column int) {
	cstr := C.CString(attribute)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_tree_view_column_add_attribute(v.Native(),
		renderer.toCellRenderer(), (*C.gchar)(cstr), C.gint(column))
}

func (v *TreeViewColumn) SetExpand(expand bool) {
	C.gtk_tree_view_column_set_expand(v.Native(), gbool(expand))
}

func (v *TreeViewColumn) GetExpand() bool {
	c := C.gtk_tree_view_column_get_expand(v.Native())
	return gobool(c)
}

func (v *TreeViewColumn) SetMinWidth(minWidth int) {
	C.gtk_tree_view_column_set_min_width(v.Native(), C.gint(minWidth))
}

func (v *TreeViewColumn) GetMinWidth() int {
	c := C.gtk_tree_view_column_get_min_width(v.Native())
	return int(c)
}

/*
 * GtkWidget
 */

type Widget struct {
	glib.InitiallyUnowned
}

type IWidget interface {
	toWidget() *C.GtkWidget
}

func (v *Widget) Native() *C.GtkWidget {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkWidget(p)
}

func (v *Widget) toWidget() *C.GtkWidget {
	if v == nil {
		return nil
	}
	return v.Native()
}

func (v *Widget) Destroy() {
	C.gtk_widget_destroy(v.Native())
}

func (v *Widget) InDestruction() bool {
	return gobool(C.gtk_widget_in_destruction(v.Native()))
}

// TODO(jrick) this may require some rethinking
func (v *Widget) Destroyed(widgetPointer **Widget) {
}

func (v *Widget) Unparent() {
	C.gtk_widget_unparent(v.Native())
}

func (v *Widget) Show() {
	C.gtk_widget_show(v.Native())
}

func (v *Widget) Hide() {
	C.gtk_widget_hide(v.Native())
}

func (v *Widget) ShowNow() {
	C.gtk_widget_show_now(v.Native())
}

func (v *Widget) ShowAll() {
	C.gtk_widget_show_all(v.Native())
}

func (v *Widget) SetNoShowAll(noShowAll bool) {
	C.gtk_widget_set_no_show_all(v.Native(), gbool(noShowAll))
}

func (v *Widget) GetNoShowAll() bool {
	c := C.gtk_widget_get_no_show_all(v.Native())
	return gobool(c)
}

func (v *Widget) Map() {
	C.gtk_widget_map(v.Native())
}

func (v *Widget) Unmap() {
	C.gtk_widget_unmap(v.Native())
}

//void gtk_widget_realize(GtkWidget *widget);
//void gtk_widget_unrealize(GtkWidget *widget);
//void gtk_widget_draw(GtkWidget *widget, cairo_t *cr);
//void gtk_widget_queue_resize(GtkWidget *widget);
//void gtk_widget_queue_resize_no_redraw(GtkWidget *widget);
//GdkFrameClock *gtk_widget_get_frame_clock(GtkWidget *widget);
//guint gtk_widget_add_tick_callback (GtkWidget *widget,
//                                    GtkTickCallback callback,
//                                    gpointer user_data,
//                                    GDestroyNotify notify);
//void gtk_widget_remove_tick_callback(GtkWidget *widget, guint id);

// TODO(jrick) GtkAllocation
func (v *Widget) SizeAllocate() {
}

// TODO(jrick) GtkAccelGroup GdkModifierType GtkAccelFlags
func (v *Widget) AddAccelerator() {
}

// TODO(jrick) GtkAccelGroup GdkModifierType
func (v *Widget) RemoveAccelerator() {
}

// TODO(jrick) GtkAccelGroup
func (v *Widget) SetAccelPath() {
}

// TODO(jrick) GList
func (v *Widget) ListAccelClosures() {
}

//gboolean gtk_widget_can_activate_accel(GtkWidget *widget, guint signal_id);

func (v *Widget) Event(event *gdk.Event) bool {
	c := C.gtk_widget_event(v.Native(),
		(*C.GdkEvent)(unsafe.Pointer(event.Native())))
	return gobool(c)
}

func (v *Widget) Activate() bool {
	return gobool(C.gtk_widget_activate(v.Native()))
}

func (v *Widget) Reparent(newParent IWidget) {
	C.gtk_widget_reparent(v.Native(), newParent.toWidget())
}

// TODO(jrick) GdkRectangle
func (v *Widget) Intersect() {
}

func (v *Widget) IsFocus() bool {
	return gobool(C.gtk_widget_is_focus(v.Native()))
}

func (v *Widget) GrabFocus() {
	C.gtk_widget_grab_focus(v.Native())
}

func (v *Widget) GrabDefault() {
	C.gtk_widget_grab_default(v.Native())
}

func (v *Widget) SetName(name string) {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_widget_set_name(v.Native(), (*C.gchar)(cstr))
}

func (v *Widget) GetName() (string, error) {
	c := C.gtk_widget_get_name(v.Native())
	if c == nil {
		return "", nilPtrErr
	}
	return C.GoString((*C.char)(c)), nil
}

func (v *Widget) SetSensitive(sensitive bool) {
	C.gtk_widget_set_sensitive(v.Native(), gbool(sensitive))
}

func (v *Widget) SetParent(parent IWidget) {
	C.gtk_widget_set_parent(v.Native(), parent.toWidget())
}

func (v *Widget) GetParent() (*Widget, error) {
	c := C.gtk_widget_get_parent(v.Native())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	w := &Widget{glib.InitiallyUnowned{obj}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return w, nil
}

func (v *Widget) SetSizeRequest(width, height int) {
	C.gtk_widget_set_size_request(v.Native(), C.gint(width), C.gint(height))
}

func (v *Widget) GetSizeRequest() (width, height int) {
	var w, h C.gint
	C.gtk_widget_get_size_request(v.Native(), &w, &h)
	return int(w), int(h)
}

func (v *Widget) SetParentWindow(parentWindow *gdk.Window) {
	C.gtk_widget_set_parent_window(v.Native(),
		(*C.GdkWindow)(unsafe.Pointer(parentWindow.Native())))
}

func (v *Widget) GetParentWindow() (*gdk.Window, error) {
	c := C.gtk_widget_get_parent_window(v.Native())
	if v == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	w := &gdk.Window{obj}
	w.Ref()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return w, nil
}

func (v *Widget) SetEvents(events int) {
	C.gtk_widget_set_events(v.Native(), C.gint(events))
}

func (v *Widget) GetEvents() int {
	return int(C.gtk_widget_get_events(v.Native()))
}

func (v *Widget) AddEvents(events int) {
	C.gtk_widget_add_events(v.Native(), C.gint(events))
}

// TODO(jrick) GdkEventMask
func (v *Widget) SetDeviceEvents() {
}

// TODO(jrick) GdkEventMask
func (v *Widget) GetDeviceEvents() {
}

// TODO(jrick) GdkEventMask
func (v *Widget) AddDeviceEvents() {
}

func (v *Widget) SetDeviceEnabled(device *gdk.Device, enabled bool) {
	C.gtk_widget_set_device_enabled(v.Native(),
		(*C.GdkDevice)(unsafe.Pointer(device.Native())), gbool(enabled))
}

func (v *Widget) GetDeviceEnabled(device *gdk.Device) bool {
	c := C.gtk_widget_get_device_enabled(v.Native(),
		(*C.GdkDevice)(unsafe.Pointer(device.Native())))
	return gobool(c)
}

func (v *Widget) GetToplevel() (*Widget, error) {
	c := C.gtk_widget_get_toplevel(v.Native())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	w := &Widget{glib.InitiallyUnowned{obj}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return w, nil
}

func (v *Widget) GetTooltipText() (string, error) {
	c := C.gtk_widget_get_tooltip_text(v.Native())
	if c == nil {
		return "", nilPtrErr
	}
	return C.GoString((*C.char)(c)), nil
}

func (v *Widget) SetTooltipText(text string) {
	cstr := C.CString(text)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_widget_set_tooltip_text(v.Native(), (*C.gchar)(cstr))
}

func (v *Widget) OverrideFont(description string) {
	cstr := C.CString(description)
	defer C.free(unsafe.Pointer(cstr))
	c := C.pango_font_description_from_string(cstr)
	C.gtk_widget_override_font(v.Native(), c)
}

func (v *Widget) GetHAlign() Align {
	c := C.gtk_widget_get_halign(v.Native())
	return Align(c)
}

func (v *Widget) SetHAlign(align Align) {
	C.gtk_widget_set_halign(v.Native(), C.GtkAlign(align))
}

func (v *Widget) GetVAlign() Align {
	c := C.gtk_widget_get_valign(v.Native())
	return Align(c)
}

func (v *Widget) SetVAlign(align Align) {
	C.gtk_widget_set_valign(v.Native(), C.GtkAlign(align))
}

func (v *Widget) GetHExpand() bool {
	c := C.gtk_widget_get_hexpand(v.Native())
	return gobool(c)
}

func (v *Widget) SetHExpand(expand bool) {
	C.gtk_widget_set_hexpand(v.Native(), gbool(expand))
}

func (v *Widget) GetVExpand() bool {
	c := C.gtk_widget_get_vexpand(v.Native())
	return gobool(c)
}

func (v *Widget) SetVExpand(expand bool) {
	C.gtk_widget_set_vexpand(v.Native(), gbool(expand))
}

// others...

/*
 * GtkWindow
 */

type Window struct {
	Bin
}

type IWindow interface {
	toWindow() *C.GtkWindow
}

func (v *Window) Native() *C.GtkWindow {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkWindow(p)
}

func (v *Window) toWindow() *C.GtkWindow {
	if v == nil {
		return nil
	}
	return v.Native()
}

func WindowNew(t WindowType) (*Window, error) {
	c := C.gtk_window_new(C.GtkWindowType(t))
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	w := &Window{Bin{Container{Widget{glib.InitiallyUnowned{obj}}}}}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return w, nil
}

func (v *Window) SetTitle(title string) {
	cstr := C.CString(title)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_window_set_title(v.Native(), (*C.gchar)(cstr))
}

func (v *Window) SetDefaultSize(width, height int) {
	C.gtk_window_set_default_size(v.Native(), C.gint(width), C.gint(height))
}

func (v *Window) SetDefaultGeometry(width, height int) {
	C.gtk_window_set_default_geometry(v.Native(), C.gint(width),
		C.gint(height))
}

// TODO(jrick) GdkGeometry GdkWindowHints
func (v *Window) SetGeometryHints() {
}

// TODO(jrick) GdkGravity
func (v *Window) SetGravity() {
}

// TODO(jrick) GdkGravity
func (v *Window) GetGravity() {
}

func (v *Window) SetPosition(position WindowPosition) {
	C.gtk_window_set_position(v.Native(), C.GtkWindowPosition(position))
}

func (v *Window) SetTransientFor(parent IWindow) {
	C.gtk_window_set_transient_for(v.Native(), parent.toWindow())
}
