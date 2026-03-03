package repository

import "testing"

func TestMockUserRepositoryImplementsInterface(t *testing.T) {
	var _ UserRepository = (*MockUserRepository)(nil)
}
