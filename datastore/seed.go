package datastore

import "context"

// This is the seedData function that is used to seed the datastore with some initial data

// seedData seeds the datastore with some initial data
func (ds *Datastore) seedItemData() {
	// Add some items to the inventory
	ds.AddItem(context.Background(), Item{
		SKU:   "item-1",
		Name:  "Item 1",
		Price: 100.00,
	}, 2)
	ds.AddItem(context.Background(), Item{
		SKU:   "item-2",
		Name:  "Item 2",
		Price: 200.00,
	}, 2)
	ds.AddItem(context.Background(), Item{
		SKU:   "item-3",
		Name:  "Item 3",
		Price: 300.00,
	}, 2)
}

// seedData seeds the datastore with some initial data
func (us *UserStore) seedData() {
	us.AddUser("test", "Test User", "test")
	us.AddUser("admin", "Admin User", "admin")
}

// seedData seeds the datastore with some initial data
func (us *SessionStore) seedData() {
	us.sessions["test"] = "test"
	us.users["test"] = User{
		ID:   "test",
		Name: "Test User",
	}
	us.sessions["admin"] = "admin"
	us.users["admin"] = User{
		ID:   "admin",
		Name: "Admin User",
	}
}
