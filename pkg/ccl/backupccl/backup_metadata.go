package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/ccl/storageccl"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/protoreflect"
	"github.com/cockroachdb/cockroach/pkg/sql/stats"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

const (
	metadataSSTName  = "metadata.sst"
	fileInfoPath     = "fileinfo.sst"
	sstBackupKey     = "backup"
	sstDescsPrefix   = "desc/"
	sstFilesPrefix   = "file/"
	sstNamesPrefix   = "name/"
	sstSpansPrefix   = "span/"
	sstStatsPrefix   = "stats/"
	sstTenantsPrefix = "tenant/"
)

func writeBackupMetadataSST(
	ctx context.Context,
	dest cloud.ExternalStorage,
	enc *jobspb.BackupEncryptionOptions,
	manifest *BackupManifest,
	stats []*stats.TableStatisticProto,
) error {
	__antithesis_instrumentation__.Notify(7326)
	var w io.WriteCloser
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		__antithesis_instrumentation__.Notify(7330)
		cancel()
		if w != nil {
			__antithesis_instrumentation__.Notify(7331)
			w.Close()
		} else {
			__antithesis_instrumentation__.Notify(7332)
		}
	}()
	__antithesis_instrumentation__.Notify(7327)

	w, err := makeWriter(ctx, dest, metadataSSTName, enc)
	if err != nil {
		__antithesis_instrumentation__.Notify(7333)
		return err
	} else {
		__antithesis_instrumentation__.Notify(7334)
	}
	__antithesis_instrumentation__.Notify(7328)

	if err := constructMetadataSST(ctx, dest, enc, w, manifest, stats); err != nil {
		__antithesis_instrumentation__.Notify(7335)
		return err
	} else {
		__antithesis_instrumentation__.Notify(7336)
	}
	__antithesis_instrumentation__.Notify(7329)

	err = w.Close()
	w = nil
	return err
}

func makeWriter(
	ctx context.Context,
	dest cloud.ExternalStorage,
	filename string,
	enc *jobspb.BackupEncryptionOptions,
) (io.WriteCloser, error) {
	__antithesis_instrumentation__.Notify(7337)
	w, err := dest.Writer(ctx, filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(7340)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(7341)
	}
	__antithesis_instrumentation__.Notify(7338)

	if enc != nil {
		__antithesis_instrumentation__.Notify(7342)
		key, err := getEncryptionKey(ctx, enc, dest.Settings(), dest.ExternalIOConf())
		if err != nil {
			__antithesis_instrumentation__.Notify(7345)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(7346)
		}
		__antithesis_instrumentation__.Notify(7343)
		encW, err := storageccl.EncryptingWriter(w, key)
		if err != nil {
			__antithesis_instrumentation__.Notify(7347)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(7348)
		}
		__antithesis_instrumentation__.Notify(7344)
		w = encW
	} else {
		__antithesis_instrumentation__.Notify(7349)
	}
	__antithesis_instrumentation__.Notify(7339)
	return w, nil
}

func constructMetadataSST(
	ctx context.Context,
	dest cloud.ExternalStorage,
	enc *jobspb.BackupEncryptionOptions,
	w io.Writer,
	m *BackupManifest,
	stats []*stats.TableStatisticProto,
) error {
	__antithesis_instrumentation__.Notify(7350)

	sst := storage.MakeBackupSSTWriter(ctx, dest.Settings(), w)
	defer sst.Close()

	if err := writeManifestToMetadata(ctx, sst, m); err != nil {
		__antithesis_instrumentation__.Notify(7358)
		return err
	} else {
		__antithesis_instrumentation__.Notify(7359)
	}
	__antithesis_instrumentation__.Notify(7351)

	if err := writeDescsToMetadata(ctx, sst, m); err != nil {
		__antithesis_instrumentation__.Notify(7360)
		return err
	} else {
		__antithesis_instrumentation__.Notify(7361)
	}
	__antithesis_instrumentation__.Notify(7352)

	if err := writeFilesToMetadata(ctx, sst, m, dest, enc, fileInfoPath); err != nil {
		__antithesis_instrumentation__.Notify(7362)
		return err
	} else {
		__antithesis_instrumentation__.Notify(7363)
	}
	__antithesis_instrumentation__.Notify(7353)

	if err := writeNamesToMetadata(ctx, sst, m); err != nil {
		__antithesis_instrumentation__.Notify(7364)
		return err
	} else {
		__antithesis_instrumentation__.Notify(7365)
	}
	__antithesis_instrumentation__.Notify(7354)

	if err := writeSpansToMetadata(ctx, sst, m); err != nil {
		__antithesis_instrumentation__.Notify(7366)
		return err
	} else {
		__antithesis_instrumentation__.Notify(7367)
	}
	__antithesis_instrumentation__.Notify(7355)

	if err := writeStatsToMetadata(ctx, sst, stats); err != nil {
		__antithesis_instrumentation__.Notify(7368)
		return err
	} else {
		__antithesis_instrumentation__.Notify(7369)
	}
	__antithesis_instrumentation__.Notify(7356)

	if err := writeTenantsToMetadata(ctx, sst, m); err != nil {
		__antithesis_instrumentation__.Notify(7370)
		return err
	} else {
		__antithesis_instrumentation__.Notify(7371)
	}
	__antithesis_instrumentation__.Notify(7357)

	return sst.Finish()
}

func writeManifestToMetadata(ctx context.Context, sst storage.SSTWriter, m *BackupManifest) error {
	__antithesis_instrumentation__.Notify(7372)
	info := *m
	info.Descriptors = nil
	info.DescriptorChanges = nil
	info.Files = nil
	info.Spans = nil
	info.StatisticsFilenames = nil
	info.IntroducedSpans = nil
	info.Tenants = nil

	b, err := protoutil.Marshal(&info)
	if err != nil {
		__antithesis_instrumentation__.Notify(7374)
		return err
	} else {
		__antithesis_instrumentation__.Notify(7375)
	}
	__antithesis_instrumentation__.Notify(7373)
	return sst.PutUnversioned(roachpb.Key(sstBackupKey), b)
}

func writeDescsToMetadata(ctx context.Context, sst storage.SSTWriter, m *BackupManifest) error {
	__antithesis_instrumentation__.Notify(7376)

	if len(m.DescriptorChanges) > 0 {
		__antithesis_instrumentation__.Notify(7378)
		sort.Slice(m.DescriptorChanges, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(7380)
			if m.DescriptorChanges[i].ID < m.DescriptorChanges[j].ID {
				__antithesis_instrumentation__.Notify(7382)
				return true
			} else {
				__antithesis_instrumentation__.Notify(7383)
				if m.DescriptorChanges[i].ID == m.DescriptorChanges[j].ID {
					__antithesis_instrumentation__.Notify(7384)
					return !m.DescriptorChanges[i].Time.Less(m.DescriptorChanges[j].Time)
				} else {
					__antithesis_instrumentation__.Notify(7385)
				}
			}
			__antithesis_instrumentation__.Notify(7381)
			return false
		})
		__antithesis_instrumentation__.Notify(7379)
		for _, i := range m.DescriptorChanges {
			__antithesis_instrumentation__.Notify(7386)
			k := encodeDescSSTKey(i.ID)
			var b []byte
			if i.Desc != nil {
				__antithesis_instrumentation__.Notify(7388)
				t, _, _, _ := descpb.FromDescriptor(i.Desc)
				if t == nil || func() bool {
					__antithesis_instrumentation__.Notify(7389)
					return !t.Dropped() == true
				}() == true {
					__antithesis_instrumentation__.Notify(7390)
					bytes, err := protoutil.Marshal(i.Desc)
					if err != nil {
						__antithesis_instrumentation__.Notify(7392)
						return err
					} else {
						__antithesis_instrumentation__.Notify(7393)
					}
					__antithesis_instrumentation__.Notify(7391)
					b = bytes
				} else {
					__antithesis_instrumentation__.Notify(7394)
				}
			} else {
				__antithesis_instrumentation__.Notify(7395)
			}
			__antithesis_instrumentation__.Notify(7387)
			if err := sst.PutMVCC(storage.MVCCKey{Key: k, Timestamp: i.Time}, b); err != nil {
				__antithesis_instrumentation__.Notify(7396)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7397)
			}

		}
	} else {
		__antithesis_instrumentation__.Notify(7398)
		sort.Slice(m.Descriptors, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(7400)
			return descID(m.Descriptors[i]) < descID(m.Descriptors[j])
		})
		__antithesis_instrumentation__.Notify(7399)
		for _, i := range m.Descriptors {
			__antithesis_instrumentation__.Notify(7401)
			id := descID(i)
			k := encodeDescSSTKey(id)
			b, err := protoutil.Marshal(&i)
			if err != nil {
				__antithesis_instrumentation__.Notify(7403)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7404)
			}
			__antithesis_instrumentation__.Notify(7402)

			if m.StartTime.IsEmpty() {
				__antithesis_instrumentation__.Notify(7405)
				if err := sst.PutUnversioned(k, b); err != nil {
					__antithesis_instrumentation__.Notify(7406)
					return err
				} else {
					__antithesis_instrumentation__.Notify(7407)
				}
			} else {
				__antithesis_instrumentation__.Notify(7408)
				if err := sst.PutMVCC(storage.MVCCKey{Key: k, Timestamp: m.StartTime}, b); err != nil {
					__antithesis_instrumentation__.Notify(7409)
					return err
				} else {
					__antithesis_instrumentation__.Notify(7410)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(7377)
	return nil
}

func writeFilesToMetadata(
	ctx context.Context,
	sst storage.SSTWriter,
	m *BackupManifest,
	dest cloud.ExternalStorage,
	enc *jobspb.BackupEncryptionOptions,
	fileInfoPath string,
) error {
	__antithesis_instrumentation__.Notify(7411)
	w, err := makeWriter(ctx, dest, fileInfoPath, enc)
	if err != nil {
		__antithesis_instrumentation__.Notify(7417)
		return err
	} else {
		__antithesis_instrumentation__.Notify(7418)
	}
	__antithesis_instrumentation__.Notify(7412)
	defer w.Close()
	fileSST := storage.MakeBackupSSTWriter(ctx, dest.Settings(), w)
	defer fileSST.Close()

	sort.Slice(m.Files, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(7419)
		cmp := m.Files[i].Span.Key.Compare(m.Files[j].Span.Key)
		return cmp < 0 || func() bool {
			__antithesis_instrumentation__.Notify(7420)
			return (cmp == 0 && func() bool {
				__antithesis_instrumentation__.Notify(7421)
				return strings.Compare(m.Files[i].Path, m.Files[j].Path) < 0 == true
			}() == true) == true
		}() == true
	})
	__antithesis_instrumentation__.Notify(7413)

	for _, i := range m.Files {
		__antithesis_instrumentation__.Notify(7422)
		b, err := protoutil.Marshal(&i)
		if err != nil {
			__antithesis_instrumentation__.Notify(7424)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7425)
		}
		__antithesis_instrumentation__.Notify(7423)
		if err := fileSST.PutUnversioned(encodeFileSSTKey(i.Span.Key, i.Path), b); err != nil {
			__antithesis_instrumentation__.Notify(7426)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7427)
		}
	}
	__antithesis_instrumentation__.Notify(7414)

	err = fileSST.Finish()
	if err != nil {
		__antithesis_instrumentation__.Notify(7428)
		return err
	} else {
		__antithesis_instrumentation__.Notify(7429)
	}
	__antithesis_instrumentation__.Notify(7415)
	err = w.Close()
	if err != nil {
		__antithesis_instrumentation__.Notify(7430)
		return err
	} else {
		__antithesis_instrumentation__.Notify(7431)
	}
	__antithesis_instrumentation__.Notify(7416)

	return sst.PutUnversioned(encodeFilenameSSTKey(fileInfoPath), nil)
}

type name struct {
	parent, parentSchema descpb.ID
	name                 string
	id                   descpb.ID
	ts                   hlc.Timestamp
}

type namespace []name

func (a namespace) Len() int { __antithesis_instrumentation__.Notify(7432); return len(a) }
func (a namespace) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(7433)
	a[i], a[j] = a[j], a[i]
}
func (a namespace) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(7434)
	if a[i].parent == a[j].parent {
		__antithesis_instrumentation__.Notify(7436)
		if a[i].parentSchema == a[j].parentSchema {
			__antithesis_instrumentation__.Notify(7438)
			cmp := strings.Compare(a[i].name, a[j].name)
			return cmp < 0 || func() bool {
				__antithesis_instrumentation__.Notify(7439)
				return (cmp == 0 && func() bool {
					__antithesis_instrumentation__.Notify(7440)
					return (a[i].ts.IsEmpty() || func() bool {
						__antithesis_instrumentation__.Notify(7441)
						return a[j].ts.Less(a[i].ts) == true
					}() == true) == true
				}() == true) == true
			}() == true
		} else {
			__antithesis_instrumentation__.Notify(7442)
		}
		__antithesis_instrumentation__.Notify(7437)
		return a[i].parentSchema < a[j].parentSchema
	} else {
		__antithesis_instrumentation__.Notify(7443)
	}
	__antithesis_instrumentation__.Notify(7435)
	return a[i].parent < a[j].parent
}

func writeNamesToMetadata(ctx context.Context, sst storage.SSTWriter, m *BackupManifest) error {
	__antithesis_instrumentation__.Notify(7444)
	revs := m.DescriptorChanges
	if len(revs) == 0 {
		__antithesis_instrumentation__.Notify(7448)
		revs = make([]BackupManifest_DescriptorRevision, len(m.Descriptors))
		for i := range m.Descriptors {
			__antithesis_instrumentation__.Notify(7449)
			revs[i].Desc = &m.Descriptors[i]
			revs[i].Time = m.EndTime
			revs[i].ID = descID(m.Descriptors[i])
		}
	} else {
		__antithesis_instrumentation__.Notify(7450)
	}
	__antithesis_instrumentation__.Notify(7445)

	names := make(namespace, len(revs))

	for i, rev := range revs {
		__antithesis_instrumentation__.Notify(7451)
		names[i].id = rev.ID
		names[i].ts = rev.Time
		tb, db, typ, sc := descpb.FromDescriptor(rev.Desc)
		if db != nil {
			__antithesis_instrumentation__.Notify(7452)
			names[i].name = db.Name
		} else {
			__antithesis_instrumentation__.Notify(7453)
			if sc != nil {
				__antithesis_instrumentation__.Notify(7454)
				names[i].name = sc.Name
				names[i].parent = sc.ParentID
			} else {
				__antithesis_instrumentation__.Notify(7455)
				if tb != nil {
					__antithesis_instrumentation__.Notify(7456)
					names[i].name = tb.Name
					names[i].parent = tb.ParentID
					names[i].parentSchema = keys.PublicSchemaID
					if s := tb.UnexposedParentSchemaID; s != descpb.InvalidID {
						__antithesis_instrumentation__.Notify(7458)
						names[i].parentSchema = s
					} else {
						__antithesis_instrumentation__.Notify(7459)
					}
					__antithesis_instrumentation__.Notify(7457)
					if tb.Dropped() {
						__antithesis_instrumentation__.Notify(7460)
						names[i].id = 0
					} else {
						__antithesis_instrumentation__.Notify(7461)
					}
				} else {
					__antithesis_instrumentation__.Notify(7462)
					if typ != nil {
						__antithesis_instrumentation__.Notify(7463)
						names[i].name = typ.Name
						names[i].parent = typ.ParentID
						names[i].parentSchema = typ.ParentSchemaID
					} else {
						__antithesis_instrumentation__.Notify(7464)
					}
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(7446)
	sort.Sort(names)

	for i, rev := range names {
		__antithesis_instrumentation__.Notify(7465)
		if i > 0 {
			__antithesis_instrumentation__.Notify(7467)
			prev := names[i-1]
			prev.ts = rev.ts
			if prev == rev {
				__antithesis_instrumentation__.Notify(7468)
				continue
			} else {
				__antithesis_instrumentation__.Notify(7469)
			}
		} else {
			__antithesis_instrumentation__.Notify(7470)
		}
		__antithesis_instrumentation__.Notify(7466)
		k := encodeNameSSTKey(rev.parent, rev.parentSchema, rev.name)
		v := encoding.EncodeUvarintAscending(nil, uint64(rev.id))
		if err := sst.PutMVCC(storage.MVCCKey{Key: k, Timestamp: rev.ts}, v); err != nil {
			__antithesis_instrumentation__.Notify(7471)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7472)
		}
	}
	__antithesis_instrumentation__.Notify(7447)

	return nil
}

func writeSpansToMetadata(ctx context.Context, sst storage.SSTWriter, m *BackupManifest) error {
	__antithesis_instrumentation__.Notify(7473)
	sort.Sort(roachpb.Spans(m.Spans))
	sort.Sort(roachpb.Spans(m.IntroducedSpans))

	for i, j := 0, 0; i < len(m.Spans) || func() bool {
		__antithesis_instrumentation__.Notify(7475)
		return j < len(m.IntroducedSpans) == true
	}() == true; {
		__antithesis_instrumentation__.Notify(7476)
		var sp roachpb.Span
		var ts hlc.Timestamp

		if j >= len(m.IntroducedSpans) {
			__antithesis_instrumentation__.Notify(7478)
			sp = m.Spans[i]
			ts = m.StartTime
			i++
		} else {
			__antithesis_instrumentation__.Notify(7479)
			if i >= len(m.Spans) {
				__antithesis_instrumentation__.Notify(7480)
				sp = m.IntroducedSpans[j]
				ts = hlc.Timestamp{}
				j++
			} else {
				__antithesis_instrumentation__.Notify(7481)
				cmp := m.Spans[i].Key.Compare(m.IntroducedSpans[j].Key)
				if cmp < 0 {
					__antithesis_instrumentation__.Notify(7482)
					sp = m.Spans[i]
					ts = m.StartTime
					i++
				} else {
					__antithesis_instrumentation__.Notify(7483)
					sp = m.IntroducedSpans[j]
					ts = hlc.Timestamp{}
					j++
				}
			}
		}
		__antithesis_instrumentation__.Notify(7477)
		if ts.IsEmpty() {
			__antithesis_instrumentation__.Notify(7484)
			if err := sst.PutUnversioned(encodeSpanSSTKey(sp), nil); err != nil {
				__antithesis_instrumentation__.Notify(7485)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7486)
			}
		} else {
			__antithesis_instrumentation__.Notify(7487)
			k := storage.MVCCKey{Key: encodeSpanSSTKey(sp), Timestamp: ts}
			if err := sst.PutMVCC(k, nil); err != nil {
				__antithesis_instrumentation__.Notify(7488)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7489)
			}
		}
	}
	__antithesis_instrumentation__.Notify(7474)
	return nil
}

func writeStatsToMetadata(
	ctx context.Context, sst storage.SSTWriter, stats []*stats.TableStatisticProto,
) error {
	__antithesis_instrumentation__.Notify(7490)
	sort.Slice(stats, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(7493)
		return stats[i].TableID < stats[j].TableID || func() bool {
			__antithesis_instrumentation__.Notify(7494)
			return (stats[i].TableID == stats[j].TableID && func() bool {
				__antithesis_instrumentation__.Notify(7495)
				return stats[i].StatisticID < stats[j].StatisticID == true
			}() == true) == true
		}() == true
	})
	__antithesis_instrumentation__.Notify(7491)

	for _, i := range stats {
		__antithesis_instrumentation__.Notify(7496)
		b, err := protoutil.Marshal(i)
		if err != nil {
			__antithesis_instrumentation__.Notify(7498)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7499)
		}
		__antithesis_instrumentation__.Notify(7497)
		if err := sst.PutUnversioned(encodeStatSSTKey(i.TableID, i.StatisticID), b); err != nil {
			__antithesis_instrumentation__.Notify(7500)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7501)
		}
	}
	__antithesis_instrumentation__.Notify(7492)
	return nil
}

func writeTenantsToMetadata(ctx context.Context, sst storage.SSTWriter, m *BackupManifest) error {
	__antithesis_instrumentation__.Notify(7502)
	sort.Slice(m.Tenants, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(7505)
		return m.Tenants[i].ID < m.Tenants[j].ID
	})
	__antithesis_instrumentation__.Notify(7503)
	for _, i := range m.Tenants {
		__antithesis_instrumentation__.Notify(7506)
		b, err := protoutil.Marshal(&i)
		if err != nil {
			__antithesis_instrumentation__.Notify(7508)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7509)
		}
		__antithesis_instrumentation__.Notify(7507)
		if err := sst.PutUnversioned(encodeTenantSSTKey(i.ID), b); err != nil {
			__antithesis_instrumentation__.Notify(7510)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7511)
		}
	}
	__antithesis_instrumentation__.Notify(7504)
	return nil
}

func descID(in descpb.Descriptor) descpb.ID {
	__antithesis_instrumentation__.Notify(7512)
	switch i := in.Union.(type) {
	case *descpb.Descriptor_Table:
		__antithesis_instrumentation__.Notify(7513)
		return i.Table.ID
	case *descpb.Descriptor_Database:
		__antithesis_instrumentation__.Notify(7514)
		return i.Database.ID
	case *descpb.Descriptor_Type:
		__antithesis_instrumentation__.Notify(7515)
		return i.Type.ID
	case *descpb.Descriptor_Schema:
		__antithesis_instrumentation__.Notify(7516)
		return i.Schema.ID
	default:
		__antithesis_instrumentation__.Notify(7517)
		panic(fmt.Sprintf("unknown desc %T", in))
	}
}

func deprefix(key roachpb.Key, prefix string) (roachpb.Key, error) {
	__antithesis_instrumentation__.Notify(7518)
	if !bytes.HasPrefix(key, []byte(prefix)) {
		__antithesis_instrumentation__.Notify(7520)
		return nil, errors.Errorf("malformed key missing expected prefix %s: %q", prefix, key)
	} else {
		__antithesis_instrumentation__.Notify(7521)
	}
	__antithesis_instrumentation__.Notify(7519)
	return key[len(prefix):], nil
}

func encodeDescSSTKey(id descpb.ID) roachpb.Key {
	__antithesis_instrumentation__.Notify(7522)
	return roachpb.Key(encoding.EncodeUvarintAscending([]byte(sstDescsPrefix), uint64(id)))
}

func decodeDescSSTKey(key roachpb.Key) (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(7523)
	key, err := deprefix(key, sstDescsPrefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(7525)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(7526)
	}
	__antithesis_instrumentation__.Notify(7524)
	_, id, err := encoding.DecodeUvarintAscending(key)
	return descpb.ID(id), err
}

func encodeFileSSTKey(spanStart roachpb.Key, filename string) roachpb.Key {
	__antithesis_instrumentation__.Notify(7527)
	buf := make([]byte, 0)
	buf = encoding.EncodeBytesAscending(buf, spanStart)
	return roachpb.Key(encoding.EncodeStringAscending(buf, filename))
}

func encodeFilenameSSTKey(filename string) roachpb.Key {
	__antithesis_instrumentation__.Notify(7528)
	return encoding.EncodeStringAscending([]byte(sstFilesPrefix), filename)
}

func decodeUnsafeFileSSTKey(key roachpb.Key) (roachpb.Key, string, error) {
	__antithesis_instrumentation__.Notify(7529)
	key, spanStart, err := encoding.DecodeBytesAscending(key, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(7532)
		return nil, "", err
	} else {
		__antithesis_instrumentation__.Notify(7533)
	}
	__antithesis_instrumentation__.Notify(7530)
	_, filename, err := encoding.DecodeUnsafeStringAscending(key, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(7534)
		return nil, "", err
	} else {
		__antithesis_instrumentation__.Notify(7535)
	}
	__antithesis_instrumentation__.Notify(7531)
	return roachpb.Key(spanStart), filename, err
}

func decodeUnsafeFileInfoSSTKey(key roachpb.Key) (string, error) {
	__antithesis_instrumentation__.Notify(7536)
	key, err := deprefix(key, sstFilesPrefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(7539)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(7540)
	}
	__antithesis_instrumentation__.Notify(7537)

	_, path, err := encoding.DecodeUnsafeStringAscending(key, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(7541)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(7542)
	}
	__antithesis_instrumentation__.Notify(7538)
	return path, err
}

func encodeNameSSTKey(parentDB, parentSchema descpb.ID, name string) roachpb.Key {
	__antithesis_instrumentation__.Notify(7543)
	buf := []byte(sstNamesPrefix)
	buf = encoding.EncodeUvarintAscending(buf, uint64(parentDB))
	buf = encoding.EncodeUvarintAscending(buf, uint64(parentSchema))
	return roachpb.Key(encoding.EncodeStringAscending(buf, name))
}

func decodeUnsafeNameSSTKey(key roachpb.Key) (descpb.ID, descpb.ID, string, error) {
	__antithesis_instrumentation__.Notify(7544)
	key, err := deprefix(key, sstNamesPrefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(7549)
		return 0, 0, "", err
	} else {
		__antithesis_instrumentation__.Notify(7550)
	}
	__antithesis_instrumentation__.Notify(7545)
	key, parentID, err := encoding.DecodeUvarintAscending(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(7551)
		return 0, 0, "", err
	} else {
		__antithesis_instrumentation__.Notify(7552)
	}
	__antithesis_instrumentation__.Notify(7546)
	key, schemaID, err := encoding.DecodeUvarintAscending(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(7553)
		return 0, 0, "", err
	} else {
		__antithesis_instrumentation__.Notify(7554)
	}
	__antithesis_instrumentation__.Notify(7547)
	_, name, err := encoding.DecodeUnsafeStringAscending(key, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(7555)
		return 0, 0, "", err
	} else {
		__antithesis_instrumentation__.Notify(7556)
	}
	__antithesis_instrumentation__.Notify(7548)
	return descpb.ID(parentID), descpb.ID(schemaID), name, nil
}

func encodeSpanSSTKey(span roachpb.Span) roachpb.Key {
	__antithesis_instrumentation__.Notify(7557)
	buf := encoding.EncodeBytesAscending([]byte(sstSpansPrefix), span.Key)
	return roachpb.Key(encoding.EncodeBytesAscending(buf, span.EndKey))
}

func decodeSpanSSTKey(key roachpb.Key) (roachpb.Span, error) {
	__antithesis_instrumentation__.Notify(7558)
	key, err := deprefix(key, sstSpansPrefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(7561)
		return roachpb.Span{}, err
	} else {
		__antithesis_instrumentation__.Notify(7562)
	}
	__antithesis_instrumentation__.Notify(7559)
	key, start, err := encoding.DecodeBytesAscending(key, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(7563)
		return roachpb.Span{}, err
	} else {
		__antithesis_instrumentation__.Notify(7564)
	}
	__antithesis_instrumentation__.Notify(7560)
	_, end, err := encoding.DecodeBytesAscending(key, nil)
	return roachpb.Span{Key: start, EndKey: end}, err
}

func encodeStatSSTKey(id descpb.ID, statID uint64) roachpb.Key {
	__antithesis_instrumentation__.Notify(7565)
	buf := encoding.EncodeUvarintAscending([]byte(sstStatsPrefix), uint64(id))
	return roachpb.Key(encoding.EncodeUvarintAscending(buf, statID))
}

func decodeStatSSTKey(key roachpb.Key) (descpb.ID, uint64, error) {
	__antithesis_instrumentation__.Notify(7566)
	key, err := deprefix(key, sstStatsPrefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(7569)
		return 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(7570)
	}
	__antithesis_instrumentation__.Notify(7567)
	key, id, err := encoding.DecodeUvarintAscending(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(7571)
		return 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(7572)
	}
	__antithesis_instrumentation__.Notify(7568)
	_, stat, err := encoding.DecodeUvarintAscending(key)
	return descpb.ID(id), stat, err
}

func encodeTenantSSTKey(id uint64) roachpb.Key {
	__antithesis_instrumentation__.Notify(7573)
	return encoding.EncodeUvarintAscending([]byte(sstTenantsPrefix), id)
}

func decodeTenantSSTKey(key roachpb.Key) (uint64, error) {
	__antithesis_instrumentation__.Notify(7574)
	key, err := deprefix(key, sstTenantsPrefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(7577)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(7578)
	}
	__antithesis_instrumentation__.Notify(7575)
	_, id, err := encoding.DecodeUvarintAscending(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(7579)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(7580)
	}
	__antithesis_instrumentation__.Notify(7576)
	return id, nil
}

func pbBytesToJSON(in []byte, msg protoutil.Message) (json.JSON, error) {
	__antithesis_instrumentation__.Notify(7581)
	if err := protoutil.Unmarshal(in, msg); err != nil {
		__antithesis_instrumentation__.Notify(7584)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(7585)
	}
	__antithesis_instrumentation__.Notify(7582)
	j, err := protoreflect.MessageToJSON(msg, protoreflect.FmtFlags{})
	if err != nil {
		__antithesis_instrumentation__.Notify(7586)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(7587)
	}
	__antithesis_instrumentation__.Notify(7583)
	return j, nil
}

func debugDumpFileSST(
	ctx context.Context,
	store cloud.ExternalStorage,
	fileInfoPath string,
	enc *jobspb.BackupEncryptionOptions,
	out func(rawKey, readableKey string, value json.JSON) error,
) error {
	__antithesis_instrumentation__.Notify(7588)
	var encOpts *roachpb.FileEncryptionOptions
	if enc != nil {
		__antithesis_instrumentation__.Notify(7592)
		key, err := getEncryptionKey(ctx, enc, store.Settings(), store.ExternalIOConf())
		if err != nil {
			__antithesis_instrumentation__.Notify(7594)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7595)
		}
		__antithesis_instrumentation__.Notify(7593)
		encOpts = &roachpb.FileEncryptionOptions{Key: key}
	} else {
		__antithesis_instrumentation__.Notify(7596)
	}
	__antithesis_instrumentation__.Notify(7589)
	iter, err := storageccl.ExternalSSTReader(ctx, store, fileInfoPath, encOpts)
	if err != nil {
		__antithesis_instrumentation__.Notify(7597)
		return err
	} else {
		__antithesis_instrumentation__.Notify(7598)
	}
	__antithesis_instrumentation__.Notify(7590)
	defer iter.Close()
	for iter.SeekGE(storage.MVCCKey{}); ; iter.Next() {
		__antithesis_instrumentation__.Notify(7599)
		ok, err := iter.Valid()
		if err != nil {
			__antithesis_instrumentation__.Notify(7604)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7605)
		}
		__antithesis_instrumentation__.Notify(7600)
		if !ok {
			__antithesis_instrumentation__.Notify(7606)
			break
		} else {
			__antithesis_instrumentation__.Notify(7607)
		}
		__antithesis_instrumentation__.Notify(7601)
		k := iter.UnsafeKey()
		spanStart, path, err := decodeUnsafeFileSSTKey(k.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(7608)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7609)
		}
		__antithesis_instrumentation__.Notify(7602)
		f, err := pbBytesToJSON(iter.UnsafeValue(), &BackupManifest_File{})
		if err != nil {
			__antithesis_instrumentation__.Notify(7610)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7611)
		}
		__antithesis_instrumentation__.Notify(7603)
		if err := out(k.String(), fmt.Sprintf("file %s (%s)", path, spanStart.String()), f); err != nil {
			__antithesis_instrumentation__.Notify(7612)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7613)
		}
	}
	__antithesis_instrumentation__.Notify(7591)

	return nil
}

func DebugDumpMetadataSST(
	ctx context.Context,
	store cloud.ExternalStorage,
	path string,
	enc *jobspb.BackupEncryptionOptions,
	out func(rawKey, readableKey string, value json.JSON) error,
) error {
	__antithesis_instrumentation__.Notify(7614)
	var encOpts *roachpb.FileEncryptionOptions
	if enc != nil {
		__antithesis_instrumentation__.Notify(7618)
		key, err := getEncryptionKey(ctx, enc, store.Settings(), store.ExternalIOConf())
		if err != nil {
			__antithesis_instrumentation__.Notify(7620)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7621)
		}
		__antithesis_instrumentation__.Notify(7619)
		encOpts = &roachpb.FileEncryptionOptions{Key: key}
	} else {
		__antithesis_instrumentation__.Notify(7622)
	}
	__antithesis_instrumentation__.Notify(7615)

	iter, err := storageccl.ExternalSSTReader(ctx, store, path, encOpts)
	if err != nil {
		__antithesis_instrumentation__.Notify(7623)
		return err
	} else {
		__antithesis_instrumentation__.Notify(7624)
	}
	__antithesis_instrumentation__.Notify(7616)
	defer iter.Close()

	for iter.SeekGE(storage.MVCCKey{}); ; iter.Next() {
		__antithesis_instrumentation__.Notify(7625)
		ok, err := iter.Valid()
		if err != nil {
			__antithesis_instrumentation__.Notify(7628)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7629)
		}
		__antithesis_instrumentation__.Notify(7626)
		if !ok {
			__antithesis_instrumentation__.Notify(7630)
			break
		} else {
			__antithesis_instrumentation__.Notify(7631)
		}
		__antithesis_instrumentation__.Notify(7627)
		k := iter.UnsafeKey()
		switch {
		case bytes.Equal(k.Key, []byte(sstBackupKey)):
			__antithesis_instrumentation__.Notify(7632)
			info, err := pbBytesToJSON(iter.UnsafeValue(), &BackupManifest{})
			if err != nil {
				__antithesis_instrumentation__.Notify(7652)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7653)
			}
			__antithesis_instrumentation__.Notify(7633)
			if err := out(k.String(), "backup info", info); err != nil {
				__antithesis_instrumentation__.Notify(7654)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7655)
			}

		case bytes.HasPrefix(k.Key, []byte(sstDescsPrefix)):
			__antithesis_instrumentation__.Notify(7634)
			id, err := decodeDescSSTKey(k.Key)
			if err != nil {
				__antithesis_instrumentation__.Notify(7656)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7657)
			}
			__antithesis_instrumentation__.Notify(7635)
			var desc json.JSON
			if v := iter.UnsafeValue(); len(v) > 0 {
				__antithesis_instrumentation__.Notify(7658)
				desc, err = pbBytesToJSON(v, &descpb.Descriptor{})
				if err != nil {
					__antithesis_instrumentation__.Notify(7659)
					return err
				} else {
					__antithesis_instrumentation__.Notify(7660)
				}
			} else {
				__antithesis_instrumentation__.Notify(7661)
			}
			__antithesis_instrumentation__.Notify(7636)
			if err := out(k.String(), fmt.Sprintf("desc %d @ %v", id, k.Timestamp), desc); err != nil {
				__antithesis_instrumentation__.Notify(7662)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7663)
			}

		case bytes.HasPrefix(k.Key, []byte(sstFilesPrefix)):
			__antithesis_instrumentation__.Notify(7637)
			p, err := decodeUnsafeFileInfoSSTKey(k.Key)
			if err != nil {
				__antithesis_instrumentation__.Notify(7664)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7665)
			}
			__antithesis_instrumentation__.Notify(7638)
			if err := out(k.String(), fmt.Sprintf("file info @ %s", p), nil); err != nil {
				__antithesis_instrumentation__.Notify(7666)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7667)
			}
			__antithesis_instrumentation__.Notify(7639)
			if err := debugDumpFileSST(ctx, store, p, enc, out); err != nil {
				__antithesis_instrumentation__.Notify(7668)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7669)
			}
		case bytes.HasPrefix(k.Key, []byte(sstNamesPrefix)):
			__antithesis_instrumentation__.Notify(7640)
			db, sc, name, err := decodeUnsafeNameSSTKey(k.Key)
			if err != nil {
				__antithesis_instrumentation__.Notify(7670)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7671)
			}
			__antithesis_instrumentation__.Notify(7641)
			var id uint64
			if v := iter.UnsafeValue(); len(v) > 0 {
				__antithesis_instrumentation__.Notify(7672)
				_, id, err = encoding.DecodeUvarintAscending(v)
				if err != nil {
					__antithesis_instrumentation__.Notify(7673)
					return err
				} else {
					__antithesis_instrumentation__.Notify(7674)
				}
			} else {
				__antithesis_instrumentation__.Notify(7675)
			}
			__antithesis_instrumentation__.Notify(7642)
			mapping := fmt.Sprintf("name db %d / schema %d / %q @ %v -> %d", db, sc, name, k.Timestamp, id)
			if err := out(k.String(), mapping, nil); err != nil {
				__antithesis_instrumentation__.Notify(7676)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7677)
			}

		case bytes.HasPrefix(k.Key, []byte(sstSpansPrefix)):
			__antithesis_instrumentation__.Notify(7643)
			span, err := decodeSpanSSTKey(k.Key)
			if err != nil {
				__antithesis_instrumentation__.Notify(7678)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7679)
			}
			__antithesis_instrumentation__.Notify(7644)
			if err := out(k.String(), fmt.Sprintf("span %s @ %v", span, k.Timestamp), nil); err != nil {
				__antithesis_instrumentation__.Notify(7680)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7681)
			}

		case bytes.HasPrefix(k.Key, []byte(sstStatsPrefix)):
			__antithesis_instrumentation__.Notify(7645)
			tblID, statID, err := decodeStatSSTKey(k.Key)
			if err != nil {
				__antithesis_instrumentation__.Notify(7682)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7683)
			}
			__antithesis_instrumentation__.Notify(7646)
			s, err := pbBytesToJSON(iter.UnsafeValue(), &stats.TableStatisticProto{})
			if err != nil {
				__antithesis_instrumentation__.Notify(7684)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7685)
			}
			__antithesis_instrumentation__.Notify(7647)
			if err := out(k.String(), fmt.Sprintf("stats tbl %d, id %d", tblID, statID), s); err != nil {
				__antithesis_instrumentation__.Notify(7686)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7687)
			}

		case bytes.HasPrefix(k.Key, []byte(sstTenantsPrefix)):
			__antithesis_instrumentation__.Notify(7648)
			id, err := decodeTenantSSTKey(k.Key)
			if err != nil {
				__antithesis_instrumentation__.Notify(7688)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7689)
			}
			__antithesis_instrumentation__.Notify(7649)
			i, err := pbBytesToJSON(iter.UnsafeValue(), &descpb.TenantInfo{})
			if err != nil {
				__antithesis_instrumentation__.Notify(7690)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7691)
			}
			__antithesis_instrumentation__.Notify(7650)
			if err := out(k.String(), fmt.Sprintf("tenant %d", id), i); err != nil {
				__antithesis_instrumentation__.Notify(7692)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7693)
			}

		default:
			__antithesis_instrumentation__.Notify(7651)
			if err := out(k.String(), "unknown", json.FromString(fmt.Sprintf("%q", iter.UnsafeValue()))); err != nil {
				__antithesis_instrumentation__.Notify(7694)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7695)
			}
		}
	}
	__antithesis_instrumentation__.Notify(7617)

	return nil
}

type BackupMetadata struct {
	BackupManifest
	store    cloud.ExternalStorage
	enc      *jobspb.BackupEncryptionOptions
	filename string
}

func newBackupMetadata(
	ctx context.Context,
	exportStore cloud.ExternalStorage,
	sstFileName string,
	encryption *jobspb.BackupEncryptionOptions,
) (*BackupMetadata, error) {
	__antithesis_instrumentation__.Notify(7696)
	var encOpts *roachpb.FileEncryptionOptions
	if encryption != nil {
		__antithesis_instrumentation__.Notify(7702)
		key, err := getEncryptionKey(ctx, encryption, exportStore.Settings(), exportStore.ExternalIOConf())
		if err != nil {
			__antithesis_instrumentation__.Notify(7704)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(7705)
		}
		__antithesis_instrumentation__.Notify(7703)
		encOpts = &roachpb.FileEncryptionOptions{Key: key}
	} else {
		__antithesis_instrumentation__.Notify(7706)
	}
	__antithesis_instrumentation__.Notify(7697)

	iter, err := storageccl.ExternalSSTReader(ctx, exportStore, sstFileName, encOpts)
	if err != nil {
		__antithesis_instrumentation__.Notify(7707)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(7708)
	}
	__antithesis_instrumentation__.Notify(7698)
	defer iter.Close()

	var sstManifest BackupManifest
	iter.SeekGE(storage.MakeMVCCMetadataKey([]byte(sstBackupKey)))
	ok, err := iter.Valid()
	if err != nil {
		__antithesis_instrumentation__.Notify(7709)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(7710)
	}
	__antithesis_instrumentation__.Notify(7699)
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(7711)
		return !iter.UnsafeKey().Key.Equal([]byte(sstBackupKey)) == true
	}() == true {
		__antithesis_instrumentation__.Notify(7712)
		return nil, errors.Errorf("metadata SST does not contain backup manifest")
	} else {
		__antithesis_instrumentation__.Notify(7713)
	}
	__antithesis_instrumentation__.Notify(7700)

	if err := protoutil.Unmarshal(iter.UnsafeValue(), &sstManifest); err != nil {
		__antithesis_instrumentation__.Notify(7714)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(7715)
	}
	__antithesis_instrumentation__.Notify(7701)

	return &BackupMetadata{BackupManifest: sstManifest, store: exportStore, enc: encryption, filename: sstFileName}, nil
}

type SpanIterator struct {
	backing bytesIter
	filter  func(key storage.MVCCKey) bool
	err     error
}

func (b *BackupMetadata) SpanIter(ctx context.Context) SpanIterator {
	__antithesis_instrumentation__.Notify(7716)
	backing := makeBytesIter(ctx, b.store, b.filename, []byte(sstSpansPrefix), b.enc, true)
	return SpanIterator{
		backing: backing,
	}
}

func (b *BackupMetadata) IntroducedSpanIter(ctx context.Context) SpanIterator {
	__antithesis_instrumentation__.Notify(7717)
	backing := makeBytesIter(ctx, b.store, b.filename, []byte(sstSpansPrefix), b.enc, false)

	return SpanIterator{
		backing: backing,
		filter: func(key storage.MVCCKey) bool {
			__antithesis_instrumentation__.Notify(7718)
			return key.Timestamp == hlc.Timestamp{}
		},
	}
}

func (si *SpanIterator) Close() {
	__antithesis_instrumentation__.Notify(7719)
	si.backing.close()
}

func (si *SpanIterator) Err() error {
	__antithesis_instrumentation__.Notify(7720)
	if si.err != nil {
		__antithesis_instrumentation__.Notify(7722)
		return si.err
	} else {
		__antithesis_instrumentation__.Notify(7723)
	}
	__antithesis_instrumentation__.Notify(7721)
	return si.backing.err()
}

func (si *SpanIterator) Next(span *roachpb.Span) bool {
	__antithesis_instrumentation__.Notify(7724)
	wrapper := resultWrapper{}

	for si.backing.next(&wrapper) {
		__antithesis_instrumentation__.Notify(7726)
		if si.filter == nil || func() bool {
			__antithesis_instrumentation__.Notify(7727)
			return si.filter(wrapper.key) == true
		}() == true {
			__antithesis_instrumentation__.Notify(7728)
			sp, err := decodeSpanSSTKey(wrapper.key.Key)
			if err != nil {
				__antithesis_instrumentation__.Notify(7730)
				si.err = err
				return false
			} else {
				__antithesis_instrumentation__.Notify(7731)
			}
			__antithesis_instrumentation__.Notify(7729)

			*span = sp
			return true
		} else {
			__antithesis_instrumentation__.Notify(7732)
		}
	}
	__antithesis_instrumentation__.Notify(7725)

	return false
}

type FileIterator struct {
	mergedIterator   storage.SimpleMVCCIterator
	backingIterators []storage.SimpleMVCCIterator
	err              error
}

func (b *BackupMetadata) FileIter(ctx context.Context) FileIterator {
	__antithesis_instrumentation__.Notify(7733)
	fileInfoIter := makeBytesIter(ctx, b.store, b.filename, []byte(sstFilesPrefix), b.enc, false)
	defer fileInfoIter.close()

	var iters []storage.SimpleMVCCIterator
	var encOpts *roachpb.FileEncryptionOptions
	if b.enc != nil {
		__antithesis_instrumentation__.Notify(7737)
		key, err := getEncryptionKey(ctx, b.enc, b.store.Settings(), b.store.ExternalIOConf())
		if err != nil {
			__antithesis_instrumentation__.Notify(7739)
			return FileIterator{err: err}
		} else {
			__antithesis_instrumentation__.Notify(7740)
		}
		__antithesis_instrumentation__.Notify(7738)
		encOpts = &roachpb.FileEncryptionOptions{Key: key}
	} else {
		__antithesis_instrumentation__.Notify(7741)
	}
	__antithesis_instrumentation__.Notify(7734)

	result := resultWrapper{}
	for fileInfoIter.next(&result) {
		__antithesis_instrumentation__.Notify(7742)
		path, err := decodeUnsafeFileInfoSSTKey(result.key.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(7745)
			break
		} else {
			__antithesis_instrumentation__.Notify(7746)
		}
		__antithesis_instrumentation__.Notify(7743)

		iter, err := storageccl.ExternalSSTReader(ctx, b.store, path, encOpts)
		if err != nil {
			__antithesis_instrumentation__.Notify(7747)
			return FileIterator{err: err}
		} else {
			__antithesis_instrumentation__.Notify(7748)
		}
		__antithesis_instrumentation__.Notify(7744)
		iters = append(iters, iter)
	}
	__antithesis_instrumentation__.Notify(7735)

	if fileInfoIter.err() != nil {
		__antithesis_instrumentation__.Notify(7749)
		return FileIterator{err: fileInfoIter.err()}
	} else {
		__antithesis_instrumentation__.Notify(7750)
	}
	__antithesis_instrumentation__.Notify(7736)

	mergedIter := storage.MakeMultiIterator(iters)
	mergedIter.SeekGE(storage.MVCCKey{})
	return FileIterator{mergedIterator: mergedIter, backingIterators: iters}
}

func (fi *FileIterator) Close() {
	__antithesis_instrumentation__.Notify(7751)
	for _, it := range fi.backingIterators {
		__antithesis_instrumentation__.Notify(7753)
		it.Close()
	}
	__antithesis_instrumentation__.Notify(7752)
	fi.mergedIterator = nil
	fi.backingIterators = fi.backingIterators[:0]
}

func (fi *FileIterator) Err() error {
	__antithesis_instrumentation__.Notify(7754)
	return fi.err
}

func (fi *FileIterator) Next(file *BackupManifest_File) bool {
	__antithesis_instrumentation__.Notify(7755)
	if fi.err != nil {
		__antithesis_instrumentation__.Notify(7759)
		return false
	} else {
		__antithesis_instrumentation__.Notify(7760)
	}
	__antithesis_instrumentation__.Notify(7756)

	valid, err := fi.mergedIterator.Valid()
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(7761)
		return !valid == true
	}() == true {
		__antithesis_instrumentation__.Notify(7762)
		fi.err = err
		return false
	} else {
		__antithesis_instrumentation__.Notify(7763)
	}
	__antithesis_instrumentation__.Notify(7757)
	err = protoutil.Unmarshal(fi.mergedIterator.UnsafeValue(), file)
	if err != nil {
		__antithesis_instrumentation__.Notify(7764)
		fi.err = err
		return false
	} else {
		__antithesis_instrumentation__.Notify(7765)
	}
	__antithesis_instrumentation__.Notify(7758)

	fi.mergedIterator.Next()
	return true
}

type DescIterator struct {
	backing bytesIter
	err     error
}

func (b *BackupMetadata) DescIter(ctx context.Context) DescIterator {
	__antithesis_instrumentation__.Notify(7766)
	backing := makeBytesIter(ctx, b.store, b.filename, []byte(sstDescsPrefix), b.enc, true)
	return DescIterator{
		backing: backing,
	}
}

func (di *DescIterator) Close() {
	__antithesis_instrumentation__.Notify(7767)
	di.backing.close()
}

func (di *DescIterator) Err() error {
	__antithesis_instrumentation__.Notify(7768)
	if di.err != nil {
		__antithesis_instrumentation__.Notify(7770)
		return di.err
	} else {
		__antithesis_instrumentation__.Notify(7771)
	}
	__antithesis_instrumentation__.Notify(7769)
	return di.backing.err()
}

func (di *DescIterator) Next(desc *descpb.Descriptor) bool {
	__antithesis_instrumentation__.Notify(7772)
	wrapper := resultWrapper{}

	for di.backing.next(&wrapper) {
		__antithesis_instrumentation__.Notify(7774)
		err := protoutil.Unmarshal(wrapper.value, desc)
		if err != nil {
			__antithesis_instrumentation__.Notify(7776)
			di.err = err
			return false
		} else {
			__antithesis_instrumentation__.Notify(7777)
		}
		__antithesis_instrumentation__.Notify(7775)

		tbl, db, typ, sc := descpb.FromDescriptor(desc)
		if tbl != nil || func() bool {
			__antithesis_instrumentation__.Notify(7778)
			return db != nil == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(7779)
			return typ != nil == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(7780)
			return sc != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(7781)
			return true
		} else {
			__antithesis_instrumentation__.Notify(7782)
		}
	}
	__antithesis_instrumentation__.Notify(7773)

	return false
}

type TenantIterator struct {
	backing bytesIter
	err     error
}

func (b *BackupMetadata) TenantIter(ctx context.Context) TenantIterator {
	__antithesis_instrumentation__.Notify(7783)
	backing := makeBytesIter(ctx, b.store, b.filename, []byte(sstTenantsPrefix), b.enc, false)
	return TenantIterator{
		backing: backing,
	}
}

func (ti *TenantIterator) Close() {
	__antithesis_instrumentation__.Notify(7784)
	ti.backing.close()
}

func (ti *TenantIterator) Err() error {
	__antithesis_instrumentation__.Notify(7785)
	if ti.err != nil {
		__antithesis_instrumentation__.Notify(7787)
		return ti.err
	} else {
		__antithesis_instrumentation__.Notify(7788)
	}
	__antithesis_instrumentation__.Notify(7786)
	return ti.backing.err()
}

func (ti *TenantIterator) Next(tenant *descpb.TenantInfoWithUsage) bool {
	__antithesis_instrumentation__.Notify(7789)
	wrapper := resultWrapper{}
	ok := ti.backing.next(&wrapper)
	if !ok {
		__antithesis_instrumentation__.Notify(7792)
		return false
	} else {
		__antithesis_instrumentation__.Notify(7793)
	}
	__antithesis_instrumentation__.Notify(7790)

	err := protoutil.Unmarshal(wrapper.value, tenant)
	if err != nil {
		__antithesis_instrumentation__.Notify(7794)
		ti.err = err
		return false
	} else {
		__antithesis_instrumentation__.Notify(7795)
	}
	__antithesis_instrumentation__.Notify(7791)

	return true
}

type DescriptorRevisionIterator struct {
	backing bytesIter
	err     error
}

func (b *BackupMetadata) DescriptorChangesIter(ctx context.Context) DescriptorRevisionIterator {
	__antithesis_instrumentation__.Notify(7796)
	backing := makeBytesIter(ctx, b.store, b.filename, []byte(sstDescsPrefix), b.enc, false)
	return DescriptorRevisionIterator{
		backing: backing,
	}
}

func (dri *DescriptorRevisionIterator) Close() {
	__antithesis_instrumentation__.Notify(7797)
	dri.backing.close()
}

func (dri *DescriptorRevisionIterator) Err() error {
	__antithesis_instrumentation__.Notify(7798)
	if dri.err != nil {
		__antithesis_instrumentation__.Notify(7800)
		return dri.err
	} else {
		__antithesis_instrumentation__.Notify(7801)
	}
	__antithesis_instrumentation__.Notify(7799)
	return dri.backing.err()
}

func (dri *DescriptorRevisionIterator) Next(revision *BackupManifest_DescriptorRevision) bool {
	__antithesis_instrumentation__.Notify(7802)
	wrapper := resultWrapper{}
	ok := dri.backing.next(&wrapper)
	if !ok {
		__antithesis_instrumentation__.Notify(7805)
		return false
	} else {
		__antithesis_instrumentation__.Notify(7806)
	}
	__antithesis_instrumentation__.Notify(7803)

	err := unmarshalWrapper(&wrapper, revision)
	if err != nil {
		__antithesis_instrumentation__.Notify(7807)
		dri.err = err
		return false
	} else {
		__antithesis_instrumentation__.Notify(7808)
	}
	__antithesis_instrumentation__.Notify(7804)

	return true
}

func unmarshalWrapper(wrapper *resultWrapper, rev *BackupManifest_DescriptorRevision) error {
	__antithesis_instrumentation__.Notify(7809)
	var desc *descpb.Descriptor
	if len(wrapper.value) > 0 {
		__antithesis_instrumentation__.Notify(7812)
		desc = &descpb.Descriptor{}
		err := protoutil.Unmarshal(wrapper.value, desc)
		if err != nil {
			__antithesis_instrumentation__.Notify(7813)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7814)
		}
	} else {
		__antithesis_instrumentation__.Notify(7815)
	}
	__antithesis_instrumentation__.Notify(7810)

	id, err := decodeDescSSTKey(wrapper.key.Key)
	if err != nil {
		__antithesis_instrumentation__.Notify(7816)
		return err
	} else {
		__antithesis_instrumentation__.Notify(7817)
	}
	__antithesis_instrumentation__.Notify(7811)

	*rev = BackupManifest_DescriptorRevision{
		Desc: desc,
		ID:   id,
		Time: wrapper.key.Timestamp,
	}
	return nil
}

type StatsIterator struct {
	backing bytesIter
	err     error
}

func (b *BackupMetadata) StatsIter(ctx context.Context) StatsIterator {
	__antithesis_instrumentation__.Notify(7818)
	backing := makeBytesIter(ctx, b.store, b.filename, []byte(sstStatsPrefix), b.enc, false)
	return StatsIterator{
		backing: backing,
	}
}

func (si *StatsIterator) Close() {
	__antithesis_instrumentation__.Notify(7819)
	si.backing.close()
}

func (si *StatsIterator) Err() error {
	__antithesis_instrumentation__.Notify(7820)
	if si.err != nil {
		__antithesis_instrumentation__.Notify(7822)
		return si.err
	} else {
		__antithesis_instrumentation__.Notify(7823)
	}
	__antithesis_instrumentation__.Notify(7821)
	return si.backing.err()
}

func (si *StatsIterator) Next(statsPtr **stats.TableStatisticProto) bool {
	__antithesis_instrumentation__.Notify(7824)
	wrapper := resultWrapper{}
	ok := si.backing.next(&wrapper)

	if !ok {
		__antithesis_instrumentation__.Notify(7827)
		return false
	} else {
		__antithesis_instrumentation__.Notify(7828)
	}
	__antithesis_instrumentation__.Notify(7825)

	var s stats.TableStatisticProto
	err := protoutil.Unmarshal(wrapper.value, &s)
	if err != nil {
		__antithesis_instrumentation__.Notify(7829)
		si.err = err
		return false
	} else {
		__antithesis_instrumentation__.Notify(7830)
	}
	__antithesis_instrumentation__.Notify(7826)

	*statsPtr = &s
	return true
}

type bytesIter struct {
	Iter storage.SimpleMVCCIterator

	prefix      []byte
	useMVCCNext bool
	iterError   error
}

func makeBytesIter(
	ctx context.Context,
	store cloud.ExternalStorage,
	path string,
	prefix []byte,
	enc *jobspb.BackupEncryptionOptions,
	useMVCCNext bool,
) bytesIter {
	__antithesis_instrumentation__.Notify(7831)
	var encOpts *roachpb.FileEncryptionOptions
	if enc != nil {
		__antithesis_instrumentation__.Notify(7834)
		key, err := getEncryptionKey(ctx, enc, store.Settings(), store.ExternalIOConf())
		if err != nil {
			__antithesis_instrumentation__.Notify(7836)
			return bytesIter{iterError: err}
		} else {
			__antithesis_instrumentation__.Notify(7837)
		}
		__antithesis_instrumentation__.Notify(7835)
		encOpts = &roachpb.FileEncryptionOptions{Key: key}
	} else {
		__antithesis_instrumentation__.Notify(7838)
	}
	__antithesis_instrumentation__.Notify(7832)

	iter, err := storageccl.ExternalSSTReader(ctx, store, path, encOpts)
	if err != nil {
		__antithesis_instrumentation__.Notify(7839)
		return bytesIter{iterError: err}
	} else {
		__antithesis_instrumentation__.Notify(7840)
	}
	__antithesis_instrumentation__.Notify(7833)

	iter.SeekGE(storage.MakeMVCCMetadataKey(prefix))
	return bytesIter{
		Iter:        iter,
		prefix:      prefix,
		useMVCCNext: useMVCCNext,
	}
}

func (bi *bytesIter) next(resWrapper *resultWrapper) bool {
	__antithesis_instrumentation__.Notify(7841)
	if bi.iterError != nil {
		__antithesis_instrumentation__.Notify(7845)
		return false
	} else {
		__antithesis_instrumentation__.Notify(7846)
	}
	__antithesis_instrumentation__.Notify(7842)

	valid, err := bi.Iter.Valid()
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(7847)
		return !valid == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(7848)
		return !bytes.HasPrefix(bi.Iter.UnsafeKey().Key, bi.prefix) == true
	}() == true {
		__antithesis_instrumentation__.Notify(7849)
		bi.close()
		bi.iterError = err
		return false
	} else {
		__antithesis_instrumentation__.Notify(7850)
	}
	__antithesis_instrumentation__.Notify(7843)

	key := bi.Iter.UnsafeKey()
	resWrapper.key.Key = key.Key.Clone()
	resWrapper.key.Timestamp = key.Timestamp
	resWrapper.value = resWrapper.value[:0]
	resWrapper.value = append(resWrapper.value, bi.Iter.UnsafeValue()...)

	if bi.useMVCCNext {
		__antithesis_instrumentation__.Notify(7851)
		bi.Iter.NextKey()
	} else {
		__antithesis_instrumentation__.Notify(7852)
		bi.Iter.Next()
	}
	__antithesis_instrumentation__.Notify(7844)
	return true
}

func (bi *bytesIter) err() error {
	__antithesis_instrumentation__.Notify(7853)
	return bi.iterError
}

func (bi *bytesIter) close() {
	__antithesis_instrumentation__.Notify(7854)
	if bi.Iter != nil {
		__antithesis_instrumentation__.Notify(7855)
		bi.Iter.Close()
		bi.Iter = nil
	} else {
		__antithesis_instrumentation__.Notify(7856)
	}
}

type resultWrapper struct {
	key   storage.MVCCKey
	value []byte
}
