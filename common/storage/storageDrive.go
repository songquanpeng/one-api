package storage

var storageDrives = New()

type StorageDrive interface {
	Upload(data []byte, fileName string) (string, error)
	Name() string
}

func New() *Storage {
	storageDrive := &Storage{
		drives: make(map[string]StorageDrive, 0),
	}

	return storageDrive
}

func AddStorageDrive(drives ...StorageDrive) {
	storageDrives.addDrives(drives...)
}

func (s *Storage) addDrives(drives ...StorageDrive) {
	for _, d := range drives {
		s.addDrive(d)
	}
}

func (s *Storage) addDrive(drive StorageDrive) {
	if drive != nil {
		driveName := drive.Name()
		if _, ok := s.drives[driveName]; ok {
			return
		}
		s.drives[driveName] = drive
	}
}
