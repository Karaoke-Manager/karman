// Package testdata provides utility functions for filling a database with test data.
// These functions usually take a *testing.T and a pgxutil.DB value.
// Each function inserts some amount of data into the db for further tests.
// If the insert fails, the test is aborted with an error message.
// If the insert is successful, a function may return some data that allows you to test against the inserted data.
package testdata
