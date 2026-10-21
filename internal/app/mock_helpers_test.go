package app

type mockStats struct {
	AccessReads, IdentityReads                                                int
	QueryCount, CreateCount, UploadCount, StartCount, RegisterCount, ACKCount int
	Proofs                                                                    []string
}

func (m *mockAPI) stats() mockStats {
	m.mu.Lock()
	defer m.mu.Unlock()
	return mockStats{QueryCount: m.QueryCount, CreateCount: m.CreateCount, UploadCount: m.UploadCount, StartCount: m.StartCount, RegisterCount: m.RegisterCount, ACKCount: m.ACKCount, AccessReads: m.AccessReads, IdentityReads: m.IdentityReads, Proofs: append([]string(nil), m.Proofs...)}
}
func (m *mockAPI) set(fn func(*mockAPI)) { m.mu.Lock(); defer m.mu.Unlock(); fn(m) }
