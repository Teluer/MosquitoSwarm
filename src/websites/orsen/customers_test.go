package orsen

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"mosquitoSwarm/src/db/dao"
	"strings"
	"testing"
)

func Test_createRandomCustomer(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		phone     string
	}{
		{
			name:      "customer",
			firstName: "Aaaa",
			lastName:  "Bbbb",
			phone:     "12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDb := new(MockDatabase)
			mockDb.On("GetLeastUsedPhone").Return(tt.phone)
			mockDb.On("GetLeastUsedFirstName").Return(tt.firstName)
			mockDb.On("GetLeastUsedLastName").Return(tt.lastName)
			dao.Dao = mockDb

			possiblePhones := []string{"0" + tt.phone, "+373" + tt.phone, "+373 " + tt.phone}
			possibleNames := []string{tt.firstName, strings.ToLower(tt.firstName),
				tt.firstName + " " + tt.lastName, strings.ToLower(tt.firstName + " " + tt.lastName),
				tt.lastName + " " + tt.firstName, strings.ToLower(tt.lastName + " " + tt.firstName),
			}

			gotName, gotPhone := CreateRandomCustomer(mockDb, "0;+373 ;+373")
			mockDb.AssertNumberOfCalls(t, "GetLeastUsedPhone", 1)
			mockDb.AssertNumberOfCalls(t, "GetLeastUsedFirstName", 1)
			assert.Contains(t, possibleNames, gotName, "CreateRandomCustomer() gotName = %v, not in %v", gotName, possibleNames)
			assert.Contains(t, possiblePhones, gotPhone, "CreateRandomCustomer() gotPhone = %v, want %v", gotPhone, possiblePhones)
		})
	}
}

type MockDatabase struct {
	dao.Database
	mock.Mock
}

func (m *MockDatabase) GetLeastUsedPhone() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDatabase) GetLeastUsedFirstName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDatabase) GetLeastUsedLastName() string {
	args := m.Called()
	return args.String(0)
}
