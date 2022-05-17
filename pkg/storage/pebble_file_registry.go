package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
	"github.com/cockroachdb/pebble/record"
	"github.com/cockroachdb/pebble/vfs"
	"github.com/cockroachdb/pebble/vfs/atomicfs"
)

var CanRegistryElideFunc func(entry *enginepb.FileEntry) bool

const maxRegistrySize = 128 << 20

type PebbleFileRegistry struct {
	FS vfs.FS

	DBDir string

	ReadOnly bool

	mu struct {
		syncutil.Mutex

		entries map[string]*enginepb.FileEntry

		registryFile vfs.File

		registryWriter *record.Writer

		marker *atomicfs.Marker

		registryFilename string
	}
}

const (
	registryFilenameBase = "COCKROACHDB_REGISTRY"
	registryMarkerName   = "registry"
)

func (r *PebbleFileRegistry) CheckNoRegistryFile() error {
	__antithesis_instrumentation__.Notify(642624)
	filename, err := atomicfs.ReadMarker(r.FS, r.DBDir, registryMarkerName)
	if oserror.IsNotExist(err) {
		__antithesis_instrumentation__.Notify(642627)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(642628)
		if err != nil {
			__antithesis_instrumentation__.Notify(642629)
			return err
		} else {
			__antithesis_instrumentation__.Notify(642630)
		}
	}
	__antithesis_instrumentation__.Notify(642625)
	if filename != "" {
		__antithesis_instrumentation__.Notify(642631)
		return oserror.ErrExist
	} else {
		__antithesis_instrumentation__.Notify(642632)
	}
	__antithesis_instrumentation__.Notify(642626)
	return nil
}

func (r *PebbleFileRegistry) Load() error {
	__antithesis_instrumentation__.Notify(642633)
	r.mu.Lock()
	defer r.mu.Unlock()

	r.mu.entries = make(map[string]*enginepb.FileEntry)

	if err := r.loadRegistryFromFile(); err != nil {
		__antithesis_instrumentation__.Notify(642636)
		return err
	} else {
		__antithesis_instrumentation__.Notify(642637)
	}
	__antithesis_instrumentation__.Notify(642634)

	if err := r.maybeElideEntries(); err != nil {
		__antithesis_instrumentation__.Notify(642638)
		return err
	} else {
		__antithesis_instrumentation__.Notify(642639)
	}
	__antithesis_instrumentation__.Notify(642635)

	return nil
}

func (r *PebbleFileRegistry) loadRegistryFromFile() error {
	__antithesis_instrumentation__.Notify(642640)

	marker, currentFilename, err := atomicfs.LocateMarker(r.FS, r.DBDir, registryMarkerName)
	if err != nil {
		__antithesis_instrumentation__.Notify(642644)
		return err
	} else {
		__antithesis_instrumentation__.Notify(642645)
	}
	__antithesis_instrumentation__.Notify(642641)
	r.mu.marker = marker

	r.mu.registryFilename = currentFilename

	if !r.ReadOnly {
		__antithesis_instrumentation__.Notify(642646)
		if err := r.mu.marker.RemoveObsolete(); err != nil {
			__antithesis_instrumentation__.Notify(642647)
			return err
		} else {
			__antithesis_instrumentation__.Notify(642648)
		}
	} else {
		__antithesis_instrumentation__.Notify(642649)
	}
	__antithesis_instrumentation__.Notify(642642)

	if _, err := r.maybeLoadExistingRegistry(); err != nil {
		__antithesis_instrumentation__.Notify(642650)
		return err
	} else {
		__antithesis_instrumentation__.Notify(642651)
	}
	__antithesis_instrumentation__.Notify(642643)
	return nil
}

func (r *PebbleFileRegistry) maybeLoadExistingRegistry() (bool, error) {
	__antithesis_instrumentation__.Notify(642652)
	if r.mu.registryFilename == "" {
		__antithesis_instrumentation__.Notify(642662)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(642663)
	}
	__antithesis_instrumentation__.Notify(642653)
	records, err := r.FS.Open(r.FS.PathJoin(r.DBDir, r.mu.registryFilename))
	if err != nil {
		__antithesis_instrumentation__.Notify(642664)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(642665)
	}
	__antithesis_instrumentation__.Notify(642654)

	rr := record.NewReader(records, 0)
	rdr, err := rr.Next()
	if err == io.EOF {
		__antithesis_instrumentation__.Notify(642666)
		return false, errors.New("pebble new file registry exists but is missing a header")
	} else {
		__antithesis_instrumentation__.Notify(642667)
	}
	__antithesis_instrumentation__.Notify(642655)
	if err != nil {
		__antithesis_instrumentation__.Notify(642668)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(642669)
	}
	__antithesis_instrumentation__.Notify(642656)
	registryHeaderBytes, err := ioutil.ReadAll(rdr)
	if err != nil {
		__antithesis_instrumentation__.Notify(642670)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(642671)
	}
	__antithesis_instrumentation__.Notify(642657)
	registryHeader := &enginepb.RegistryHeader{}
	if err := protoutil.Unmarshal(registryHeaderBytes, registryHeader); err != nil {
		__antithesis_instrumentation__.Notify(642672)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(642673)
	}
	__antithesis_instrumentation__.Notify(642658)

	if registryHeader.Version == enginepb.RegistryVersion_Base {
		__antithesis_instrumentation__.Notify(642674)
		return false, errors.New("new encryption registry with version Base should not exist")
	} else {
		__antithesis_instrumentation__.Notify(642675)
	}
	__antithesis_instrumentation__.Notify(642659)
	for {
		__antithesis_instrumentation__.Notify(642676)
		rdr, err := rr.Next()
		if err == io.EOF {
			__antithesis_instrumentation__.Notify(642681)
			break
		} else {
			__antithesis_instrumentation__.Notify(642682)
		}
		__antithesis_instrumentation__.Notify(642677)
		if err != nil {
			__antithesis_instrumentation__.Notify(642683)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(642684)
		}
		__antithesis_instrumentation__.Notify(642678)
		b, err := ioutil.ReadAll(rdr)
		if err != nil {
			__antithesis_instrumentation__.Notify(642685)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(642686)
		}
		__antithesis_instrumentation__.Notify(642679)
		batch := &enginepb.RegistryUpdateBatch{}
		if err := protoutil.Unmarshal(b, batch); err != nil {
			__antithesis_instrumentation__.Notify(642687)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(642688)
		}
		__antithesis_instrumentation__.Notify(642680)
		r.applyBatch(batch)
	}
	__antithesis_instrumentation__.Notify(642660)
	if err := records.Close(); err != nil {
		__antithesis_instrumentation__.Notify(642689)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(642690)
	}
	__antithesis_instrumentation__.Notify(642661)
	return true, nil
}

func (r *PebbleFileRegistry) maybeElideEntries() error {
	__antithesis_instrumentation__.Notify(642691)
	if r.ReadOnly {
		__antithesis_instrumentation__.Notify(642695)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(642696)
	}
	__antithesis_instrumentation__.Notify(642692)

	filenames := make([]string, 0, len(r.mu.entries))
	for filename := range r.mu.entries {
		__antithesis_instrumentation__.Notify(642697)
		filenames = append(filenames, filename)
	}
	__antithesis_instrumentation__.Notify(642693)
	sort.Strings(filenames)

	batch := &enginepb.RegistryUpdateBatch{}
	for _, filename := range filenames {
		__antithesis_instrumentation__.Notify(642698)
		entry := r.mu.entries[filename]

		if CanRegistryElideFunc != nil && func() bool {
			__antithesis_instrumentation__.Notify(642701)
			return CanRegistryElideFunc(entry) == true
		}() == true {
			__antithesis_instrumentation__.Notify(642702)
			batch.DeleteEntry(filename)
			continue
		} else {
			__antithesis_instrumentation__.Notify(642703)
		}
		__antithesis_instrumentation__.Notify(642699)

		path := filename
		if !filepath.IsAbs(path) {
			__antithesis_instrumentation__.Notify(642704)
			path = r.FS.PathJoin(r.DBDir, filename)
		} else {
			__antithesis_instrumentation__.Notify(642705)
		}
		__antithesis_instrumentation__.Notify(642700)
		if _, err := r.FS.Stat(path); oserror.IsNotExist(err) {
			__antithesis_instrumentation__.Notify(642706)
			batch.DeleteEntry(filename)
		} else {
			__antithesis_instrumentation__.Notify(642707)
		}
	}
	__antithesis_instrumentation__.Notify(642694)
	return r.processBatchLocked(batch)
}

func (r *PebbleFileRegistry) GetFileEntry(filename string) *enginepb.FileEntry {
	__antithesis_instrumentation__.Notify(642708)
	filename = r.tryMakeRelativePath(filename)
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.mu.entries[filename]
}

func (r *PebbleFileRegistry) SetFileEntry(filename string, entry *enginepb.FileEntry) error {
	__antithesis_instrumentation__.Notify(642709)

	if entry == nil {
		__antithesis_instrumentation__.Notify(642711)
		return r.MaybeDeleteEntry(filename)
	} else {
		__antithesis_instrumentation__.Notify(642712)
	}
	__antithesis_instrumentation__.Notify(642710)

	filename = r.tryMakeRelativePath(filename)

	r.mu.Lock()
	defer r.mu.Unlock()
	batch := &enginepb.RegistryUpdateBatch{}
	batch.PutEntry(filename, entry)
	return r.processBatchLocked(batch)
}

func (r *PebbleFileRegistry) MaybeDeleteEntry(filename string) error {
	__antithesis_instrumentation__.Notify(642713)
	filename = r.tryMakeRelativePath(filename)

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.mu.entries[filename] == nil {
		__antithesis_instrumentation__.Notify(642715)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(642716)
	}
	__antithesis_instrumentation__.Notify(642714)
	batch := &enginepb.RegistryUpdateBatch{}
	batch.DeleteEntry(filename)
	return r.processBatchLocked(batch)
}

func (r *PebbleFileRegistry) MaybeCopyEntry(src, dst string) error {
	__antithesis_instrumentation__.Notify(642717)
	src = r.tryMakeRelativePath(src)
	dst = r.tryMakeRelativePath(dst)

	r.mu.Lock()
	defer r.mu.Unlock()
	srcEntry := r.mu.entries[src]
	if srcEntry == nil && func() bool {
		__antithesis_instrumentation__.Notify(642720)
		return r.mu.entries[dst] == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(642721)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(642722)
	}
	__antithesis_instrumentation__.Notify(642718)
	batch := &enginepb.RegistryUpdateBatch{}
	if srcEntry == nil {
		__antithesis_instrumentation__.Notify(642723)
		batch.DeleteEntry(dst)
	} else {
		__antithesis_instrumentation__.Notify(642724)
		batch.PutEntry(dst, srcEntry)
	}
	__antithesis_instrumentation__.Notify(642719)
	return r.processBatchLocked(batch)
}

func (r *PebbleFileRegistry) MaybeLinkEntry(src, dst string) error {
	__antithesis_instrumentation__.Notify(642725)
	src = r.tryMakeRelativePath(src)
	dst = r.tryMakeRelativePath(dst)

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.mu.entries[src] == nil && func() bool {
		__antithesis_instrumentation__.Notify(642728)
		return r.mu.entries[dst] == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(642729)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(642730)
	}
	__antithesis_instrumentation__.Notify(642726)
	batch := &enginepb.RegistryUpdateBatch{}
	if r.mu.entries[src] == nil {
		__antithesis_instrumentation__.Notify(642731)
		batch.DeleteEntry(dst)
	} else {
		__antithesis_instrumentation__.Notify(642732)
		batch.PutEntry(dst, r.mu.entries[src])
	}
	__antithesis_instrumentation__.Notify(642727)
	return r.processBatchLocked(batch)
}

func (r *PebbleFileRegistry) tryMakeRelativePath(filename string) string {
	__antithesis_instrumentation__.Notify(642733)

	dbDir := r.DBDir

	if filename == dbDir {
		__antithesis_instrumentation__.Notify(642738)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(642739)
	}
	__antithesis_instrumentation__.Notify(642734)
	if len(dbDir) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(642740)
		return dbDir[len(dbDir)-1] != os.PathSeparator == true
	}() == true {
		__antithesis_instrumentation__.Notify(642741)
		dbDir = dbDir + string(os.PathSeparator)
	} else {
		__antithesis_instrumentation__.Notify(642742)
	}
	__antithesis_instrumentation__.Notify(642735)
	if !strings.HasPrefix(filename, dbDir) {
		__antithesis_instrumentation__.Notify(642743)
		return filename
	} else {
		__antithesis_instrumentation__.Notify(642744)
	}
	__antithesis_instrumentation__.Notify(642736)
	filename = filename[len(dbDir):]
	if len(filename) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(642745)
		return filename[0] == os.PathSeparator == true
	}() == true {
		__antithesis_instrumentation__.Notify(642746)
		filename = filename[1:]
	} else {
		__antithesis_instrumentation__.Notify(642747)
	}
	__antithesis_instrumentation__.Notify(642737)
	return filename
}

func (r *PebbleFileRegistry) processBatchLocked(batch *enginepb.RegistryUpdateBatch) error {
	__antithesis_instrumentation__.Notify(642748)
	if r.ReadOnly {
		__antithesis_instrumentation__.Notify(642752)
		return errors.New("cannot write file registry since db is read-only")
	} else {
		__antithesis_instrumentation__.Notify(642753)
	}
	__antithesis_instrumentation__.Notify(642749)
	if batch.Empty() {
		__antithesis_instrumentation__.Notify(642754)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(642755)
	}
	__antithesis_instrumentation__.Notify(642750)
	if err := r.writeToRegistryFile(batch); err != nil {
		__antithesis_instrumentation__.Notify(642756)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(642757)
	}
	__antithesis_instrumentation__.Notify(642751)
	r.applyBatch(batch)
	return nil
}

func (r *PebbleFileRegistry) applyBatch(batch *enginepb.RegistryUpdateBatch) {
	__antithesis_instrumentation__.Notify(642758)
	for _, update := range batch.Updates {
		__antithesis_instrumentation__.Notify(642759)
		if update.Entry == nil {
			__antithesis_instrumentation__.Notify(642760)
			delete(r.mu.entries, update.Filename)
		} else {
			__antithesis_instrumentation__.Notify(642761)
			if r.mu.entries == nil {
				__antithesis_instrumentation__.Notify(642763)
				r.mu.entries = make(map[string]*enginepb.FileEntry)
			} else {
				__antithesis_instrumentation__.Notify(642764)
			}
			__antithesis_instrumentation__.Notify(642762)
			r.mu.entries[update.Filename] = update.Entry
		}
	}
}

func (r *PebbleFileRegistry) writeToRegistryFile(batch *enginepb.RegistryUpdateBatch) error {
	__antithesis_instrumentation__.Notify(642765)

	if r.mu.registryWriter == nil {
		__antithesis_instrumentation__.Notify(642773)
		if err := r.createNewRegistryFile(); err != nil {
			__antithesis_instrumentation__.Notify(642774)
			return errors.Wrap(err, "creating new registry file")
		} else {
			__antithesis_instrumentation__.Notify(642775)
		}
	} else {
		__antithesis_instrumentation__.Notify(642776)
	}
	__antithesis_instrumentation__.Notify(642766)
	w, err := r.mu.registryWriter.Next()
	if err != nil {
		__antithesis_instrumentation__.Notify(642777)
		return err
	} else {
		__antithesis_instrumentation__.Notify(642778)
	}
	__antithesis_instrumentation__.Notify(642767)
	b, err := protoutil.Marshal(batch)
	if err != nil {
		__antithesis_instrumentation__.Notify(642779)
		return err
	} else {
		__antithesis_instrumentation__.Notify(642780)
	}
	__antithesis_instrumentation__.Notify(642768)
	if _, err := w.Write(b); err != nil {
		__antithesis_instrumentation__.Notify(642781)
		return err
	} else {
		__antithesis_instrumentation__.Notify(642782)
	}
	__antithesis_instrumentation__.Notify(642769)
	if err := r.mu.registryWriter.Flush(); err != nil {
		__antithesis_instrumentation__.Notify(642783)
		return err
	} else {
		__antithesis_instrumentation__.Notify(642784)
	}
	__antithesis_instrumentation__.Notify(642770)
	if err := r.mu.registryFile.Sync(); err != nil {
		__antithesis_instrumentation__.Notify(642785)
		return err
	} else {
		__antithesis_instrumentation__.Notify(642786)
	}
	__antithesis_instrumentation__.Notify(642771)

	if r.mu.registryWriter.Size() > maxRegistrySize {
		__antithesis_instrumentation__.Notify(642787)
		if err := r.createNewRegistryFile(); err != nil {
			__antithesis_instrumentation__.Notify(642788)
			return errors.Wrap(err, "rotating registry file")
		} else {
			__antithesis_instrumentation__.Notify(642789)
		}
	} else {
		__antithesis_instrumentation__.Notify(642790)
	}
	__antithesis_instrumentation__.Notify(642772)
	return nil
}

func makeRegistryFilename(iter uint64) string {
	__antithesis_instrumentation__.Notify(642791)
	return fmt.Sprintf("%s_%06d", registryFilenameBase, iter)
}

func (r *PebbleFileRegistry) createNewRegistryFile() error {
	__antithesis_instrumentation__.Notify(642792)

	filename := makeRegistryFilename(r.mu.marker.NextIter())
	filepath := r.FS.PathJoin(r.DBDir, filename)
	f, err := r.FS.Create(filepath)
	if err != nil {
		__antithesis_instrumentation__.Notify(642805)
		return err
	} else {
		__antithesis_instrumentation__.Notify(642806)
	}
	__antithesis_instrumentation__.Notify(642793)
	records := record.NewWriter(f)
	w, err := records.Next()
	if err != nil {
		__antithesis_instrumentation__.Notify(642807)
		return errors.CombineErrors(err, f.Close())
	} else {
		__antithesis_instrumentation__.Notify(642808)
	}
	__antithesis_instrumentation__.Notify(642794)

	errFunc := func(err error) error {
		__antithesis_instrumentation__.Notify(642809)
		err1 := records.Close()
		err2 := f.Close()
		return errors.CombineErrors(err, errors.CombineErrors(err1, err2))
	}
	__antithesis_instrumentation__.Notify(642795)

	registryHeader := &enginepb.RegistryHeader{
		Version: enginepb.RegistryVersion_Records,
	}
	b, err := protoutil.Marshal(registryHeader)
	if err != nil {
		__antithesis_instrumentation__.Notify(642810)
		return errFunc(err)
	} else {
		__antithesis_instrumentation__.Notify(642811)
	}
	__antithesis_instrumentation__.Notify(642796)
	if _, err := w.Write(b); err != nil {
		__antithesis_instrumentation__.Notify(642812)
		return errFunc(err)
	} else {
		__antithesis_instrumentation__.Notify(642813)
	}
	__antithesis_instrumentation__.Notify(642797)

	batch := &enginepb.RegistryUpdateBatch{}
	for filename, entry := range r.mu.entries {
		__antithesis_instrumentation__.Notify(642814)
		batch.PutEntry(filename, entry)
	}
	__antithesis_instrumentation__.Notify(642798)
	b, err = protoutil.Marshal(batch)
	if err != nil {
		__antithesis_instrumentation__.Notify(642815)
		return errFunc(err)
	} else {
		__antithesis_instrumentation__.Notify(642816)
	}
	__antithesis_instrumentation__.Notify(642799)
	w, err = records.Next()
	if err != nil {
		__antithesis_instrumentation__.Notify(642817)
		return errFunc(err)
	} else {
		__antithesis_instrumentation__.Notify(642818)
	}
	__antithesis_instrumentation__.Notify(642800)
	if _, err := w.Write(b); err != nil {
		__antithesis_instrumentation__.Notify(642819)
		return errFunc(err)
	} else {
		__antithesis_instrumentation__.Notify(642820)
	}
	__antithesis_instrumentation__.Notify(642801)
	if err := records.Flush(); err != nil {
		__antithesis_instrumentation__.Notify(642821)
		return errFunc(err)
	} else {
		__antithesis_instrumentation__.Notify(642822)
	}
	__antithesis_instrumentation__.Notify(642802)
	if err := f.Sync(); err != nil {
		__antithesis_instrumentation__.Notify(642823)
		return errFunc(err)
	} else {
		__antithesis_instrumentation__.Notify(642824)
	}
	__antithesis_instrumentation__.Notify(642803)

	if err := r.mu.marker.Move(filename); err != nil {
		__antithesis_instrumentation__.Notify(642825)
		return errors.Wrap(errFunc(err), "moving marker")
	} else {
		__antithesis_instrumentation__.Notify(642826)
	}

	{
		__antithesis_instrumentation__.Notify(642827)

		err = r.closeRegistry()
		if err == nil && func() bool {
			__antithesis_instrumentation__.Notify(642828)
			return r.mu.registryFilename != "" == true
		}() == true {
			__antithesis_instrumentation__.Notify(642829)
			rmErr := r.FS.Remove(r.FS.PathJoin(r.DBDir, r.mu.registryFilename))
			if rmErr != nil && func() bool {
				__antithesis_instrumentation__.Notify(642830)
				return !oserror.IsNotExist(rmErr) == true
			}() == true {
				__antithesis_instrumentation__.Notify(642831)
				err = errors.CombineErrors(err, rmErr)
			} else {
				__antithesis_instrumentation__.Notify(642832)
			}
		} else {
			__antithesis_instrumentation__.Notify(642833)
		}
	}
	__antithesis_instrumentation__.Notify(642804)

	r.mu.registryFile = f
	r.mu.registryWriter = records
	r.mu.registryFilename = filename
	return err
}

func (r *PebbleFileRegistry) getRegistryCopy() *enginepb.FileRegistry {
	__antithesis_instrumentation__.Notify(642834)
	r.mu.Lock()
	defer r.mu.Unlock()
	rv := &enginepb.FileRegistry{
		Version: enginepb.RegistryVersion_Records,
		Files:   make(map[string]*enginepb.FileEntry, len(r.mu.entries)),
	}
	for filename, entry := range r.mu.entries {
		__antithesis_instrumentation__.Notify(642836)
		ev := &enginepb.FileEntry{}
		*ev = *entry
		rv.Files[filename] = ev
	}
	__antithesis_instrumentation__.Notify(642835)
	return rv
}

func (r *PebbleFileRegistry) Close() error {
	__antithesis_instrumentation__.Notify(642837)
	r.mu.Lock()
	defer r.mu.Unlock()
	err := r.closeRegistry()
	err = errors.CombineErrors(err, r.mu.marker.Close())
	return err
}

func (r *PebbleFileRegistry) closeRegistry() error {
	__antithesis_instrumentation__.Notify(642838)
	var err1, err2 error
	if r.mu.registryWriter != nil {
		__antithesis_instrumentation__.Notify(642841)
		err1 = r.mu.registryWriter.Close()
		r.mu.registryWriter = nil
	} else {
		__antithesis_instrumentation__.Notify(642842)
	}
	__antithesis_instrumentation__.Notify(642839)
	if r.mu.registryFile != nil {
		__antithesis_instrumentation__.Notify(642843)
		err2 = r.mu.registryFile.Close()
		r.mu.registryFile = nil
	} else {
		__antithesis_instrumentation__.Notify(642844)
	}
	__antithesis_instrumentation__.Notify(642840)
	return errors.CombineErrors(err1, err2)
}
