package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func setBashCompletionFunction() {
	__antithesis_instrumentation__.Notify(42921)

	var s []string
	for _, cmd := range []*cobra.Command{createCmd, listCmd, syncCmd, gcCmd} {
		__antithesis_instrumentation__.Notify(42923)
		s = append(s, fmt.Sprintf("%s_%s", rootCmd.Name(), cmd.Name()))
	}
	__antithesis_instrumentation__.Notify(42922)
	excluded := strings.Join(s, " | ")

	rootCmd.BashCompletionFunction = fmt.Sprintf(
		`__custom_func()
{
    # only complete the 2nd arg, e.g. adminurl <foo>
    if ! [ $c -eq 2 ]; then
    	return
    fi
    
    # don't complete commands which do not accept a cluster/host arg
    case ${last_command} in
    	%s)
    		return
    		;;
    esac
    
    local hosts_out
    if hosts_out=$(roachprod cached-hosts --cluster="${cur}" 2>/dev/null); then
    		COMPREPLY=( $( compgen -W "${hosts_out[*]}" -- "$cur" ) )
    fi
}`,
		excluded,
	)
}
