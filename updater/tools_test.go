package main

/* TODO
func TestApplyWhitelistToApps_SuccessWithWhiteList(t *testing.T) {
	apps := []string{"app1", "app2", "app3"}
	whitelist := []string{"app1", "app3"}

	filteredApps, err := applyWhitelistToApps(apps, whitelist)
	assert.Nil(t, err)
	assert.Equal(t, whitelist, filteredApps)
}

func TestApplyWhitelistToApps_EmptyListMeansNoRestrictions(t *testing.T) {
	apps := []string{"app1", "app2", "app3"}
	var whitelist []string

	filteredApps, err := applyWhitelistToApps(apps, whitelist)
	assert.Nil(t, err)
	assert.Equal(t, apps, filteredApps)
}

func TestApplyWhitelistToApps_AppNotFoundError(t *testing.T) {
	apps := []string{"app1", "app2", "app3"}
	whitelist := []string{"app1", "app4"}

	_, err := applyWhitelistToApps(apps, whitelist)
	assert.NotNil(t, err)
	expectedMessage := "app 'app4' was whitelisted but not found"
	assert.Equal(t, expectedMessage, err.Error())
}
*/
