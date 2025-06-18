package svn

type MockService struct {
}

func (svc *MockService) CurrentInfo() RepoInfo {
	return RepoInfo{
		WorkingPath: "C:/Code/GitHub/textual-test/",
		RemoteURL:   "https://svn.riouxsvn.com/textual-test",
		Revision:    64,
	}
}

func (svc *MockService) FetchInfo() error {
	return nil
}

func (svc *MockService) CurrentStatus() *RepoStatus {
	return nil
}

func (svc *MockService) FetchStatus() error {
	return nil
}

func (svc *MockService) StagePath(path string) error {
	return nil
}

func (svc *MockService) UnstagePath(path string) error {
	return nil
}

func (svc *MockService) GetPath(si SectionIdx, idx int) (string, error) {
	return "", nil
}
