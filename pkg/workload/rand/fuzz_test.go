package rand

import (
	"context"
//	"encoding/hex"
	"fmt"
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"strings"
	"testing"
)

func FuzzSQL(f *testing.F) {
	defer log.Scope(f).Close(f)

	ctx := context.Background()
	t := &testing.T{}
	s, db, _ := serverutils.StartServer(t, base.TestServerArgs{})
	defer s.Stopper().Stop(ctx)
	//sqlDB := sqlutils.MakeSQLRunner(db)
	rows, err := db.Query(
		`
SELECT proname, proargtypes::regtype[]::text[] 
FROM pg_proc WHERE 
(provolatile = 'i' OR provolatile = 's') AND
proargtypes::regtype[] && array['text','bytea','name','json','jsonb']::regtype[] AND 
provariadic = 0 AND prokind != 'a' AND prokind != 'w'`)
	if err != nil {
		f.Fatal(err)
	}
	supportedTypes := []string{"text", "bytea", "name", "json", "jsonb"}

	for rows.Next() {
		var name string
		argTypesRaw := []byte{}

		if err := rows.Scan(&name, &argTypesRaw); err != nil {
			f.Fatal(err)
		}
		s := string(argTypesRaw)
		s = strings.ReplaceAll(s, "{", "")
		s = strings.ReplaceAll(s, "}", "")
		argTypes := strings.Split(s, ",")
		unsupported := false

		for _, argT := range argTypes {
			if !contains(supportedTypes, argT) {
				unsupported = true
				break
			}
		}
		if !unsupported {
			fmt.Printf("%s %s\n", name, s)
		}
		fmt.Println(name)
	}
	f.Fuzz(func(t *testing.T, a string, b string) {
		/*if len(a) > 45 || len(a) < 42 {
			t.Skip()
			return
		}*/
		//q := `676f20746573742066757a7a2076310a5b5d6279746528225c7830305c7838305c7830655c7622290a`

		//b, _ := hex.DecodeString(q)
		/*rows, err := db.Query("select decompress($1, 'gzip')", fmt.Sprintf(`x'%s'`, hex.EncodeToString(a)))
		if err == nil {
			rows.Close()
		}*/

		rows, err := db.Query("set cluster setting $1 = $2", a, b)
		if err == nil {
			rows.Close()
		}


		//rows, err := db.Query("select to_uuid(convert_from(" + q + ", 'UTF8'))")
		/*rows, err = db.Query("select to_uuid(convert_from($1, 'UTF8'))", b)
		if err == nil {
                        rows.Close()
                }*/

		/*
		if err == nil {
			rows.Close()
		} else {
			t.Fatal(err)
		}*/
		//
		//NB: t.Log is no-op because fuzz worker discards all output: root.w = io.Discard (see src/testing/fuzz.go)
		//
		// t.Log(...)
	})
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
