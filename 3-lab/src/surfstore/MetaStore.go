package surfstore

import (
	"errors"
)

type MetaStore struct {
	FileMetaMap map[string]FileMetaData
}

func (m *MetaStore) GetFileInfoMap(_ignore *bool, serverFileInfoMap *map[string]FileMetaData) error {
	// _ignore as "Args" needs to be the first argument to the RPC
	*serverFileInfoMap = m.FileMetaMap
	return nil
}

func (m *MetaStore) UpdateFile(fileMetaData *FileMetaData, latestVersion *int) error {
	fileName := fileMetaData.Filename

	// if the file exists
	if oldMetaData, ok := m.FileMetaMap[fileName]; ok {
		// but it's outdated (older version) - return err
		if fileMetaData.Version != oldMetaData.Version + 1 {
			return errors.New("Version mistmatch - Client file is outdated")
		}
	} else if fileMetaData.Version != 1 {
		// OR the file is a new file and version isn't '1'
		return errors.New("Version Error - Version of new file must be 1")
	}

	// everything is okay
	m.FileMetaMap[fileName] = *fileMetaData
	*latestVersion = fileMetaData.Version
	return nil
}

var _ MetaStoreInterface = new(MetaStore)
