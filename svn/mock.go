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

func (svc *MockService) FetchStatus() error {
	return nil
}
