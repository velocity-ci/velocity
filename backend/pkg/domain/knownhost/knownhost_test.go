package knownhost_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/velocity-ci/velocity/backend/pkg/domain/knownhost"

	"github.com/go-playground/universal-translator"
	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	govalidator "gopkg.in/go-playground/validator.v9"
)

func setup() (*gorm.DB, *govalidator.Validate, ut.Translator) {
	db := domain.NewGORMDB(":memory:")
	validator, translator := domain.NewValidator()

	return db, validator, translator
}

func TestValidNew(t *testing.T) {
	m := knownhost.NewManager(setup())

	entry := `github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==`

	kH, errs := m.New(entry)
	assert.Nil(t, errs)

	assert.NotEmpty(t, kH.UUID)
	assert.Equal(t, entry, kH.Entry)
	assert.Equal(t, []string{"github.com"}, kH.Hosts)
	assert.Equal(t, "", kH.Comment)
	assert.Equal(t, "SHA256:nThbg6kXUpJWGl7E1IGOCspRomTxdCARLviKw6E5SY8", kH.SHA256Fingerprint)
	assert.Equal(t, "16:27:ac:a5:76:28:2d:36:63:1b:56:4d:eb:df:a6:48", kH.MD5Fingerprint)
}

func TestInvalidNew(t *testing.T) {
	m := knownhost.NewManager(setup())

	entry := `github.com ssh-rsa AAAAB3NaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==`

	kH, errs := m.New(entry)
	assert.Nil(t, kH)
	assert.NotNil(t, errs)

	assert.Equal(t, []string{"entry is not a valid key!"}, errs.ErrorMap["entry"])
}

func TestSave(t *testing.T) {
	m := knownhost.NewManager(setup())

	entry := `github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==`
	kH, _ := m.New(entry)

	err := m.Save(kH)

	assert.Nil(t, err)

	assert.True(t, m.Exists(entry))
}

func TestExists(t *testing.T) {
	m := knownhost.NewManager(setup())

	entry := `github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==`
	assert.False(t, m.Exists(entry))
}

func TestList(t *testing.T) {
	m := knownhost.NewManager(setup())

	entry := `github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==`
	kH, _ := m.New(entry)
	m.Save(kH)

	q := &domain.PagingQuery{
		Limit: 3,
		Page:  1,
	}
	items, amount := m.GetAll(q)
	assert.Len(t, items, 1)
	assert.Equal(t, amount, 1)
}

func TestDelete(t *testing.T) {
	m := knownhost.NewManager(setup())
	entry := `github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==`
	kH, _ := m.New(entry)
	m.Save(kH)

	err := m.Delete(kH)
	assert.Nil(t, err)

	assert.False(t, m.Exists(entry))
}
