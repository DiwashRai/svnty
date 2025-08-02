package svn

type MockService struct {
}

func (svc *MockService) Init() {}

func (svc *MockService) CurrentInfo() RepoInfo {
	return RepoInfo{
		WorkingPath:  "C:/Code/GitHub/textual-test/",
		RemoteURL:    "https://svn.riouxsvn.com/textual-test",
		Revision:     64,
		HeadRevision: 67,
	}
}

func (svc *MockService) FetchInfo() error {
	return nil
}

func (svc *MockService) FetchHeadRevision() error {
	return nil
}

func (svc *MockService) CurrentStatus() *RepoStatus {
	return nil
}

func (svc *MockService) FetchStatus() error {
	return nil
}

func (svc *MockService) Update() error {
	return nil
}

func (svc *MockService) StagePath(path string) error {
	return nil
}

func (svc *MockService) UnstagePath(path string) error {
	return nil
}

func (svc *MockService) FetchDiff(path string) error {
	return nil
}

func (svc *MockService) GetDiff(path string) []string {
	return nil
}

func (svc *MockService) GetPathStatus(si SectionIdx, idx int) (PathStatus, error) {
	return PathStatus{}, nil
}
func (svc *MockService) CommitStaged(msg string) error {
	return nil
}

func (svc *MockService) IsOutOfDate() bool {
	info := svc.CurrentInfo()
	return info.Revision < info.HeadRevision
}
