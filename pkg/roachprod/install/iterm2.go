package install

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const splitScript = `
 if (path to frontmost application as text) does not contain "iTerm" then
        return "Can't multi-ssh without iTerm 2 as the active terminal."
 end if
 tell application "iTerm2"
        activate
        tell (create window with default profile)
            set pane_count to %d
            set col_count to round ((pane_count) ^ 0.5) rounding up
            set row_count to round ((pane_count) / col_count) rounding up
            repeat row_count - 1 times
                tell current session
                    split horizontally with default profile
                end tell
            end repeat
            set cur to row_count
            repeat row_count times
                repeat col_count - 1 times
                    if cur is equal to (pane_count) then
                        exit repeat
                    end if
                    tell current session
                        split vertically with default profile
                    end tell
                    set cur to cur + 1
                end repeat
                tell application "System Events" to tell process "iTerm2" to click menu item "Select Pane Below" of menu 1 of menu item "Select Split Pane" of menu 1 of menu bar item "Window" of menu bar 1
            end repeat
            tell current tab
                set node_list to {%s}
                repeat with i from 1 to pane_count
                    tell item i of sessions to write text "roachprod ssh %s:" & item i of node_list
                end repeat
            end tell
        end tell
    end tell
return`

func maybeSplitScreenSSHITerm2(c *SyncedCluster) (bool, error) {
	__antithesis_instrumentation__.Notify(181609)
	osascriptPath, err := exec.LookPath("osascript")
	if err != nil {
		__antithesis_instrumentation__.Notify(181612)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(181613)
	}
	__antithesis_instrumentation__.Notify(181610)
	nodeStrings := make([]string, len(c.Nodes))
	for i, nodeID := range c.Nodes {
		__antithesis_instrumentation__.Notify(181614)
		nodeStrings[i] = fmt.Sprint(nodeID)
	}
	__antithesis_instrumentation__.Notify(181611)
	script := fmt.Sprintf(splitScript, len(c.Nodes), strings.Join(nodeStrings, ","), c.Name)
	allArgs := []string{
		"osascript",
		"-e",
		script,
	}
	return true, syscall.Exec(osascriptPath, allArgs, os.Environ())
}
