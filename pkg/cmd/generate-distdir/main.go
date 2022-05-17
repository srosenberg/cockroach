package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/build/bazel"
	"github.com/cockroachdb/cockroach/pkg/build/starlarkutil"
	"github.com/google/skylark/syntax"
)

type sourceFiles struct {
	depsBzl, workspace string
}

func getSourceFiles() (*sourceFiles, error) {
	__antithesis_instrumentation__.Notify(40173)
	depsBzl, err := bazel.Runfile("DEPS.bzl")
	if err != nil {
		__antithesis_instrumentation__.Notify(40176)
		return &sourceFiles{}, err
	} else {
		__antithesis_instrumentation__.Notify(40177)
	}
	__antithesis_instrumentation__.Notify(40174)
	workspace, err := bazel.Runfile("WORKSPACE")
	if err != nil {
		__antithesis_instrumentation__.Notify(40178)
		return &sourceFiles{}, err
	} else {
		__antithesis_instrumentation__.Notify(40179)
	}
	__antithesis_instrumentation__.Notify(40175)
	return &sourceFiles{depsBzl, workspace}, nil
}

func getShasFromDepsBzl(depsBzl string, shas map[string]string) error {
	__antithesis_instrumentation__.Notify(40180)
	artifacts, err := starlarkutil.ListArtifactsInDepsBzl(depsBzl)
	if err != nil {
		__antithesis_instrumentation__.Notify(40183)
		return err
	} else {
		__antithesis_instrumentation__.Notify(40184)
	}
	__antithesis_instrumentation__.Notify(40181)
	for _, artifact := range artifacts {
		__antithesis_instrumentation__.Notify(40185)
		shas[artifact.URL] = artifact.Sha256
	}
	__antithesis_instrumentation__.Notify(40182)
	return nil
}

func getShasFromGoDownloadSdkCall(call *syntax.CallExpr, shas map[string]string) error {
	__antithesis_instrumentation__.Notify(40186)
	var urlTmpl, version string
	baseToHash := make(map[string]string)
	for _, arg := range call.Args {
		__antithesis_instrumentation__.Notify(40190)
		switch bx := arg.(type) {
		case *syntax.BinaryExpr:
			__antithesis_instrumentation__.Notify(40191)
			if bx.Op != syntax.EQ {
				__antithesis_instrumentation__.Notify(40194)
				return fmt.Errorf("unexpected binary expression Op %d", bx.Op)
			} else {
				__antithesis_instrumentation__.Notify(40195)
			}
			__antithesis_instrumentation__.Notify(40192)
			kwarg, err := starlarkutil.ExpectIdent(bx.X)
			if err != nil {
				__antithesis_instrumentation__.Notify(40196)
				return err
			} else {
				__antithesis_instrumentation__.Notify(40197)
			}
			__antithesis_instrumentation__.Notify(40193)
			if kwarg == "sdks" {
				__antithesis_instrumentation__.Notify(40198)
				switch d := bx.Y.(type) {
				case *syntax.DictExpr:
					__antithesis_instrumentation__.Notify(40199)
					for _, ex := range d.List {
						__antithesis_instrumentation__.Notify(40201)
						switch entry := ex.(type) {
						case *syntax.DictEntry:
							__antithesis_instrumentation__.Notify(40202)
							strs, err := starlarkutil.ExpectTupleOfStrings(entry.Value, 2)
							if err != nil {
								__antithesis_instrumentation__.Notify(40205)
								return err
							} else {
								__antithesis_instrumentation__.Notify(40206)
							}
							__antithesis_instrumentation__.Notify(40203)
							baseToHash[strs[0]] = strs[1]
						default:
							__antithesis_instrumentation__.Notify(40204)
							return fmt.Errorf("expected DictEntry in dictionary expression")
						}
					}
				default:
					__antithesis_instrumentation__.Notify(40200)
					return fmt.Errorf("expected dict as value of sdks kwarg")
				}
			} else {
				__antithesis_instrumentation__.Notify(40207)
				if kwarg == "urls" {
					__antithesis_instrumentation__.Notify(40208)
					urlTmpl, err = starlarkutil.ExpectSingletonStringList(bx.Y)
					if err != nil {
						__antithesis_instrumentation__.Notify(40209)
						return err
					} else {
						__antithesis_instrumentation__.Notify(40210)
					}
				} else {
					__antithesis_instrumentation__.Notify(40211)
					if kwarg == "version" {
						__antithesis_instrumentation__.Notify(40212)
						version, err = starlarkutil.ExpectLiteralString(bx.Y)
						if err != nil {
							__antithesis_instrumentation__.Notify(40213)
							return err
						} else {
							__antithesis_instrumentation__.Notify(40214)
						}
					} else {
						__antithesis_instrumentation__.Notify(40215)
					}
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(40187)
	if urlTmpl == "" || func() bool {
		__antithesis_instrumentation__.Notify(40216)
		return version == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(40217)
		return fmt.Errorf("expected both `urls` and `version` to be set")
	} else {
		__antithesis_instrumentation__.Notify(40218)
	}
	__antithesis_instrumentation__.Notify(40188)
	for basename, sha := range baseToHash {
		__antithesis_instrumentation__.Notify(40219)
		shas[strings.ReplaceAll(urlTmpl, "{}", basename)] = sha
	}
	__antithesis_instrumentation__.Notify(40189)
	return nil
}

func getShasFromNodeRepositoriesCall(call *syntax.CallExpr, shas map[string]string) error {
	__antithesis_instrumentation__.Notify(40220)
	var nodeURLTmpl, yarnURLTmpl, nodeVersion, yarnVersion, yarnSha, yarnFilename string
	nodeBaseToHash := make(map[string]string)
	for _, arg := range call.Args {
		__antithesis_instrumentation__.Notify(40224)
		switch bx := arg.(type) {
		case *syntax.BinaryExpr:
			__antithesis_instrumentation__.Notify(40225)
			if bx.Op != syntax.EQ {
				__antithesis_instrumentation__.Notify(40228)
				return fmt.Errorf("unexpected binary expression Op %d", bx.Op)
			} else {
				__antithesis_instrumentation__.Notify(40229)
			}
			__antithesis_instrumentation__.Notify(40226)
			kwarg, err := starlarkutil.ExpectIdent(bx.X)
			if err != nil {
				__antithesis_instrumentation__.Notify(40230)
				return err
			} else {
				__antithesis_instrumentation__.Notify(40231)
			}
			__antithesis_instrumentation__.Notify(40227)
			if kwarg == "node_repositories" {
				__antithesis_instrumentation__.Notify(40232)
				switch d := bx.Y.(type) {
				case *syntax.DictExpr:
					__antithesis_instrumentation__.Notify(40233)
					for _, ex := range d.List {
						__antithesis_instrumentation__.Notify(40235)
						switch entry := ex.(type) {
						case *syntax.DictEntry:
							__antithesis_instrumentation__.Notify(40236)
							strs, err := starlarkutil.ExpectTupleOfStrings(entry.Value, 3)
							if err != nil {
								__antithesis_instrumentation__.Notify(40239)
								return err
							} else {
								__antithesis_instrumentation__.Notify(40240)
							}
							__antithesis_instrumentation__.Notify(40237)
							nodeBaseToHash[strs[0]] = strs[2]
						default:
							__antithesis_instrumentation__.Notify(40238)
							return fmt.Errorf("expected DictEntry in dictionary expression")
						}
					}
				default:
					__antithesis_instrumentation__.Notify(40234)
					return fmt.Errorf("expected dict as value of node_repositories kwarg")
				}
			} else {
				__antithesis_instrumentation__.Notify(40241)
				if kwarg == "node_urls" {
					__antithesis_instrumentation__.Notify(40242)
					nodeURLTmpl, err = starlarkutil.ExpectSingletonStringList(bx.Y)
					if err != nil {
						__antithesis_instrumentation__.Notify(40243)
						return err
					} else {
						__antithesis_instrumentation__.Notify(40244)
					}
				} else {
					__antithesis_instrumentation__.Notify(40245)
					if kwarg == "node_version" {
						__antithesis_instrumentation__.Notify(40246)
						nodeVersion, err = starlarkutil.ExpectLiteralString(bx.Y)
						if err != nil {
							__antithesis_instrumentation__.Notify(40247)
							return err
						} else {
							__antithesis_instrumentation__.Notify(40248)
						}
					} else {
						__antithesis_instrumentation__.Notify(40249)
						if kwarg == "yarn_repositories" {
							__antithesis_instrumentation__.Notify(40250)
							switch d := bx.Y.(type) {
							case *syntax.DictExpr:
								__antithesis_instrumentation__.Notify(40251)
								if len(d.List) != 1 {
									__antithesis_instrumentation__.Notify(40254)
									return fmt.Errorf("expected only one version in yarn_repositories dict")
								} else {
									__antithesis_instrumentation__.Notify(40255)
								}
								__antithesis_instrumentation__.Notify(40252)
								switch entry := d.List[0].(type) {
								case *syntax.DictEntry:
									__antithesis_instrumentation__.Notify(40256)
									strs, err := starlarkutil.ExpectTupleOfStrings(entry.Value, 3)
									if err != nil {
										__antithesis_instrumentation__.Notify(40259)
										return err
									} else {
										__antithesis_instrumentation__.Notify(40260)
									}
									__antithesis_instrumentation__.Notify(40257)
									yarnFilename = strs[0]
									yarnSha = strs[2]
								default:
									__antithesis_instrumentation__.Notify(40258)
									return fmt.Errorf("expected DictEntry in dictionary expression")
								}
							default:
								__antithesis_instrumentation__.Notify(40253)
								return fmt.Errorf("expected dict as value of yarn_repositories kwarg")
							}
						} else {
							__antithesis_instrumentation__.Notify(40261)
							if kwarg == "yarn_urls" {
								__antithesis_instrumentation__.Notify(40262)
								yarnURLTmpl, err = starlarkutil.ExpectSingletonStringList(bx.Y)
								if err != nil {
									__antithesis_instrumentation__.Notify(40263)
									return err
								} else {
									__antithesis_instrumentation__.Notify(40264)
								}
							} else {
								__antithesis_instrumentation__.Notify(40265)
								if kwarg == "yarn_version" {
									__antithesis_instrumentation__.Notify(40266)
									yarnVersion, err = starlarkutil.ExpectLiteralString(bx.Y)
									if err != nil {
										__antithesis_instrumentation__.Notify(40267)
										return err
									} else {
										__antithesis_instrumentation__.Notify(40268)
									}
								} else {
									__antithesis_instrumentation__.Notify(40269)
								}
							}
						}
					}
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(40221)
	if nodeURLTmpl == "" || func() bool {
		__antithesis_instrumentation__.Notify(40270)
		return yarnURLTmpl == "" == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(40271)
		return nodeVersion == "" == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(40272)
		return yarnVersion == "" == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(40273)
		return yarnSha == "" == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(40274)
		return yarnFilename == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(40275)
		return fmt.Errorf("did not parse all needed data from node_repositories call")
	} else {
		__antithesis_instrumentation__.Notify(40276)
	}
	__antithesis_instrumentation__.Notify(40222)
	for base, sha := range nodeBaseToHash {
		__antithesis_instrumentation__.Notify(40277)
		shas[strings.ReplaceAll(strings.ReplaceAll(nodeURLTmpl, "{version}", nodeVersion), "{filename}", base)] = sha
	}
	__antithesis_instrumentation__.Notify(40223)
	shas[strings.ReplaceAll(strings.ReplaceAll(yarnURLTmpl, "{version}", yarnVersion), "{filename}", yarnFilename)] = yarnSha
	return nil
}

func getShasFromWorkspace(workspace string, shas map[string]string) error {
	__antithesis_instrumentation__.Notify(40278)
	in, err := os.Open(workspace)
	if err != nil {
		__antithesis_instrumentation__.Notify(40282)
		return err
	} else {
		__antithesis_instrumentation__.Notify(40283)
	}
	__antithesis_instrumentation__.Notify(40279)
	defer in.Close()
	parsed, err := syntax.Parse("WORKSPACE", in, 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(40284)
		return err
	} else {
		__antithesis_instrumentation__.Notify(40285)
	}
	__antithesis_instrumentation__.Notify(40280)
	for _, stmt := range parsed.Stmts {
		__antithesis_instrumentation__.Notify(40286)
		switch s := stmt.(type) {
		case *syntax.ExprStmt:
			__antithesis_instrumentation__.Notify(40287)
			switch x := s.X.(type) {
			case *syntax.CallExpr:
				__antithesis_instrumentation__.Notify(40288)
				fun := starlarkutil.GetFunctionFromCall(x)
				if fun == "http_archive" {
					__antithesis_instrumentation__.Notify(40291)
					artifact, err := starlarkutil.GetArtifactFromHTTPArchive(x)
					if err != nil {
						__antithesis_instrumentation__.Notify(40293)
						return err
					} else {
						__antithesis_instrumentation__.Notify(40294)
					}
					__antithesis_instrumentation__.Notify(40292)
					shas[artifact.URL] = artifact.Sha256
				} else {
					__antithesis_instrumentation__.Notify(40295)
				}
				__antithesis_instrumentation__.Notify(40289)
				if fun == "go_download_sdk" {
					__antithesis_instrumentation__.Notify(40296)
					if err := getShasFromGoDownloadSdkCall(x, shas); err != nil {
						__antithesis_instrumentation__.Notify(40297)
						return err
					} else {
						__antithesis_instrumentation__.Notify(40298)
					}
				} else {
					__antithesis_instrumentation__.Notify(40299)
				}
				__antithesis_instrumentation__.Notify(40290)
				if fun == "node_repositories" {
					__antithesis_instrumentation__.Notify(40300)
					if err := getShasFromNodeRepositoriesCall(x, shas); err != nil {
						__antithesis_instrumentation__.Notify(40301)
						return err
					} else {
						__antithesis_instrumentation__.Notify(40302)
					}
				} else {
					__antithesis_instrumentation__.Notify(40303)
				}

			}
		}
	}
	__antithesis_instrumentation__.Notify(40281)
	return nil
}

func getShas(src *sourceFiles) (map[string]string, error) {
	__antithesis_instrumentation__.Notify(40304)
	ret := make(map[string]string)
	if err := getShasFromDepsBzl(src.depsBzl, ret); err != nil {
		__antithesis_instrumentation__.Notify(40307)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(40308)
	}
	__antithesis_instrumentation__.Notify(40305)
	if err := getShasFromWorkspace(src.workspace, ret); err != nil {
		__antithesis_instrumentation__.Notify(40309)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(40310)
	}
	__antithesis_instrumentation__.Notify(40306)
	return ret, nil
}

func dumpOutput(shas map[string]string) {
	__antithesis_instrumentation__.Notify(40311)
	fmt.Println(`# Code generated by generate-distdir. DO NOT EDIT.

DISTDIR_FILES = {`)
	urls := make([]string, 0, len(shas))
	for url := range shas {
		__antithesis_instrumentation__.Notify(40314)
		urls = append(urls, url)
	}
	__antithesis_instrumentation__.Notify(40312)
	sort.Strings(urls)
	for _, url := range urls {
		__antithesis_instrumentation__.Notify(40315)
		fmt.Printf(`    "%s": "%s",
`, url, shas[url])
	}
	__antithesis_instrumentation__.Notify(40313)
	fmt.Println("}")
}

func generateDistdir() error {
	__antithesis_instrumentation__.Notify(40316)
	src, err := getSourceFiles()
	if err != nil {
		__antithesis_instrumentation__.Notify(40319)
		return err
	} else {
		__antithesis_instrumentation__.Notify(40320)
	}
	__antithesis_instrumentation__.Notify(40317)
	shas, err := getShas(src)
	if err != nil {
		__antithesis_instrumentation__.Notify(40321)
		return err
	} else {
		__antithesis_instrumentation__.Notify(40322)
	}
	__antithesis_instrumentation__.Notify(40318)
	dumpOutput(shas)
	return nil
}

func main() {
	__antithesis_instrumentation__.Notify(40323)
	if err := generateDistdir(); err != nil {
		__antithesis_instrumentation__.Notify(40324)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(40325)
	}
}
