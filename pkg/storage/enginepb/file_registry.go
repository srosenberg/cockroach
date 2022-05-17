package enginepb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

func (b *RegistryUpdateBatch) Empty() bool {
	__antithesis_instrumentation__.Notify(633884)
	return len(b.Updates) == 0
}

func (b *RegistryUpdateBatch) PutEntry(filename string, entry *FileEntry) {
	__antithesis_instrumentation__.Notify(633885)
	b.Updates = append(b.Updates, &RegistryUpdate{Filename: filename, Entry: entry})
}

func (b *RegistryUpdateBatch) DeleteEntry(filename string) {
	__antithesis_instrumentation__.Notify(633886)
	b.Updates = append(b.Updates, &RegistryUpdate{Filename: filename, Entry: nil})
}
