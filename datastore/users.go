package datastore

import "fmt"

// userWithCredentials defines the structure of the user with credentials
type userWithCredentials struct {
	User     User
	Password string
}

// UserStore defines the structure of the session store
type UserStore struct {
	// Ideally you want these credentials to be stored in a secure and different from sessions store
	users map[UserID]userWithCredentials
}

// NewUserStore creates a new session store
func NewUserStore() *UserStore {
	us := &UserStore{
		users: make(map[UserID]userWithCredentials),
	}
	// TODO: Remove this
	us.seedData()
	return us
}

// Authenticate authenticates a user
func (cs *UserStore) Authenticate(userID UserID, password string) (User, bool) {
	// TODO: remove this as this is only for developement
	if userID == password {
		uc := userWithCredentials{
			User:     User{ID: userID, Name: userID},
			Password: password,
		}
		cs.users[userID] = uc
		return uc.User, true
	}

	creds, ok := cs.users[userID]
	if ok && creds.Password == password {
		return creds.User, true
	}
	return User{}, false
}

// AddUser adds a user to the session store
func (cs *UserStore) AddUser(userID UserID, userName string, password string) error {
	// check if user exists
	_, ok := cs.users[userID]
	if ok {
		return fmt.Errorf("user already exists")
	}
	user := User{
		ID:   userID,
		Name: userName,
	}
	creds := userWithCredentials{
		User:     user,
		Password: password,
	}
	cs.users[userID] = creds
	return nil
}

// GetUser retrieves a user from the session store based on the user ID\
func (cs *UserStore) GetUser(userID UserID) (User, error) {
	creds, ok := cs.users[userID]
	if !ok {
		return User{}, fmt.Errorf("user not found")
	}
	return creds.User, nil
}
