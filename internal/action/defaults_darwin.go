package action

var bufdefaults = map[string]string{
	// --- Cursor movement ---
	"Up":        "CursorUp",
	"Down":      "CursorDown",
	"Right":     "CursorRight",
	"Left":      "CursorLeft",
	"Home":      "StartOfTextToggle",
	"End":       "EndOfLine",
	"CtrlHome":  "CursorStart",
	"CtrlEnd":   "CursorEnd",
	"CtrlUp":    "CursorStart",
	"CtrlDown":  "CursorEnd",
	"PageUp":    "CursorPageUp",
	"PageDown":  "CursorPageDown",

	// --- Selection ---
	"ShiftUp":        "SelectUp",
	"ShiftDown":      "SelectDown",
	"ShiftLeft":      "SelectLeft",
	"ShiftRight":     "SelectRight",
	"ShiftHome":      "SelectToStartOfTextToggle",
	"ShiftEnd":       "SelectToEndOfLine",
	"CtrlShiftUp":    "SelectToStart",
	"CtrlShiftDown":  "SelectToEnd",
	"ShiftPageUp":    "SelectPageUp",
	"ShiftPageDown":  "SelectPageDown",

	// --- Word navigation (Alt = word level, like VSCode) ---
	"AltLeft":        "WordLeft",
	"AltRight":       "WordRight",
	"AltShiftLeft":   "SelectWordLeft",
	"AltShiftRight":  "SelectWordRight",
	"CtrlLeft":       "StartOfTextToggle",
	"CtrlRight":      "EndOfLine",
	"CtrlShiftLeft":  "SelectToStartOfTextToggle",
	"CtrlShiftRight": "SelectToEndOfLine",
	"CtrlBackspace":  "DeleteWordLeft",
	"Alt-{":          "ParagraphPrevious",
	"Alt-}":          "ParagraphNext",

	// --- Line manipulation (VSCode style) ---
	"AltUp":          "MoveLinesUp",
	"AltDown":        "MoveLinesDown",
	"AltShiftUp":     "DuplicateLine",
	"AltShiftDown":   "DuplicateLine",

	// --- Basic editing ---
	"Enter":              "InsertNewline",
	"CtrlEnter":          "InsertLineBelow",

	"Backspace":          "Backspace",
	"OldBackspace":       "Backspace",
	"CtrlH":              "Backspace",
	"AltBackspace":       "DeleteWordLeft",
	"Alt-CtrlH":          "DeleteWordLeft",
	"Delete":             "Delete",
	"Insert":             "ToggleOverwriteMode",
	"Tab":                "Autocomplete|IndentSelection|InsertTab",
	"Backtab":            "CycleAutocompleteBack|OutdentSelection|OutdentLine",

	// --- Clipboard (same as VSCode) ---
	"Ctrl-c": "Copy|CopyLine",
	"Ctrl-x": "Cut|CutLine",
	"Ctrl-v": "Paste",
	"Ctrl-a": "SelectAll",
	"Ctrl-k": "CutLine",

	// --- Search (VSCode style) ---
	"Ctrl-f":       "Find",
	"Ctrl-h":       "Find",
	"F3":           "FindNext",
	"ShiftF3":      "FindPrevious",


	// --- File operations (VSCode style) ---
	"Ctrl-p":     "OpenFile",
	"Ctrl-s":     "Save",
	"Ctrl-n":     "AddTab",

	// --- Undo/Redo (VSCode style) ---
	"Ctrl-z":       "Undo",
	"Ctrl-y":       "Redo",

	// --- Navigation / Go To ---
	"Ctrl-g":     "command-edit:goto ",
	"Ctrl-l":     "JumpLine",
	"Alt-[":      "DiffPrevious|CursorStart",
	"Alt-]":      "DiffNext|CursorEnd",

	// --- Tabs (VSCode style: Ctrl+W = close tab) ---
	"Ctrl-w":       "Quit",
	"Ctrl-q":       "QuitAll",
	"Alt-,":        "PreviousTab|LastTab",
	"Alt-.":        "NextTab|FirstTab",
	"CtrlPageUp":   "PreviousTab|LastTab",
	"CtrlPageDown": "NextTab|FirstTab",

	// --- Command mode ---
	"Ctrl-o":       "CommandMode",
	// --- Splits ---
	"Ctrl-e":       "NextSplit|FirstSplit",

	// --- View / Toggle ---
	"F1":             "ToggleHelp",
	"Alt-g":          "ToggleKeyMenu",
	"Ctrl-r":         "ToggleRuler",
	"Ctrl-b":         "ShellMode",
	"Ctrl-u":         "ToggleMacro",
	"Ctrl-j":         "PlayMacro",
	"Alt-r":          "ToggleDiffGutter",
	"CtrlSpace":      "ManualTrigger",

	// --- Multi-cursor ---
	"Ctrl-d":           "SpawnMultiCursor",
	"Alt-m":            "SpawnMultiCursorSelect",
	"Alt-n":            "RemoveAllMultiCursors",
	"Alt-p":            "RemoveMultiCursor",
	"Alt-x":            "SkipMultiCursor",

	// --- Comment (VSCode style) ---
	"Ctrl-/":         "lua:comment.comment",

	// --- Integration with file managers ---
	"F2":  "Save",
	"F4":  "Quit",
	"F7":  "Find",
	"F10": "Quit",

	// --- General ---
	"Esc": "Escape,Deselect,ClearInfo,RemoveAllMultiCursors,UnhighlightSearch",

	// --- Mouse bindings ---
	"MouseWheelUp":     "ScrollUp",
	"MouseWheelDown":   "ScrollDown",
	"MouseLeft":        "MousePress",
	"MouseLeftDrag":    "MouseDrag",
	"MouseLeftRelease": "MouseRelease",
	"MouseMiddle":      "PastePrimary",
	"Ctrl-MouseLeft":   "MouseMultiCursor",
}

var infodefaults = map[string]string{
	"Up":             "HistoryUp",
	"Down":           "HistoryDown",
	"Right":          "CursorRight",
	"Left":           "CursorLeft",
	"ShiftUp":        "SelectUp",
	"ShiftDown":      "SelectDown",
	"ShiftLeft":      "SelectLeft",
	"ShiftRight":     "SelectRight",
	"AltLeft":        "WordLeft",
	"AltRight":       "WordRight",
	"AltUp":          "CursorStart",
	"AltDown":        "CursorEnd",
	"AltShiftRight":  "SelectWordRight",
	"AltShiftLeft":   "SelectWordLeft",
	"CtrlLeft":       "StartOfTextToggle",
	"CtrlRight":      "EndOfLine",
	"CtrlShiftLeft":  "SelectToStartOfTextToggle",
	"ShiftHome":      "SelectToStartOfTextToggle",
	"CtrlShiftRight": "SelectToEndOfLine",
	"ShiftEnd":       "SelectToEndOfLine",
	"CtrlUp":         "CursorStart",
	"CtrlDown":       "CursorEnd",
	"CtrlShiftUp":    "SelectToStart",
	"CtrlShiftDown":  "SelectToEnd",
	"Enter":          "ExecuteCommand",
	"CtrlH":          "Backspace",
	"Backspace":      "Backspace",
	"OldBackspace":   "Backspace",
	"Alt-CtrlH":      "DeleteWordLeft",
	"AltBackspace":   "DeleteWordLeft",
	"Tab":            "CommandComplete",
	"Backtab":        "CycleAutocompleteBack",
	"Ctrl-z":         "Undo",
	"Ctrl-y":         "Redo",
	"Ctrl-c":         "Copy",
	"Ctrl-x":         "Cut",
	"Ctrl-k":         "CutLine",
	"Ctrl-v":         "Paste",
	"Home":           "StartOfTextToggle",
	"End":            "EndOfLine",
	"CtrlHome":       "CursorStart",
	"CtrlEnd":        "CursorEnd",
	"Delete":         "Delete",
	"Ctrl-q":         "AbortCommand",
	"Ctrl-e":         "EndOfLine",
	"Ctrl-a":         "StartOfLine",
	"Ctrl-w":         "DeleteWordLeft",
	"Insert":         "ToggleOverwriteMode",
	"Ctrl-b":         "WordLeft",
	"Ctrl-f":         "WordRight",
	"Ctrl-d":         "DeleteWordLeft",
	"Ctrl-m":         "ExecuteCommand",
	"Ctrl-n":         "HistoryDown",
	"Ctrl-p":         "HistoryUp",
	"Ctrl-u":         "SelectToStart",

	// Integration with file managers
	"F10": "AbortCommand",
	"Esc": "AbortCommand",

	// Mouse bindings
	"MouseWheelUp":     "HistoryUp",
	"MouseWheelDown":   "HistoryDown",
	"MouseLeft":        "MousePress",
	"MouseLeftDrag":    "MouseDrag",
	"MouseLeftRelease": "MouseRelease",
	"MouseMiddle":      "PastePrimary",
}
