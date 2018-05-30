package content

import (
	"reflect"

	"github.com/jinzhu/gorm"
	"github.com/ovh/lhasa/api/hateoas"
	"github.com/ovh/lhasa/api/v1"
)

const (
	defaultPageSize = 20
)

// Repository is a repository manager for applications
type Repository struct {
	db *gorm.DB
}

// NewRepository creates an application repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// GetType returns the entity type managed by this repository
func (repo *Repository) GetType() reflect.Type {
	return reflect.TypeOf(v1.Content{})
}

// GetNewEntityInstance returns a new empty instance of the entity managed by this repository
func (repo *Repository) GetNewEntityInstance() hateoas.Entity {
	return &v1.Content{}
}

// FindAll returns all entities of the repository type
func (repo *Repository) FindAll() (interface{}, error) {
	return repo.FindBy(map[string]interface{}{})
}

// FindAllPage returns a page of matching entities
func (repo *Repository) FindAllPage(pageable hateoas.Pageable) (hateoas.Page, error) {
	return repo.FindPageBy(pageable, map[string]interface{}{})
}

// FindPageBy returns a page of matching entities
func (repo *Repository) FindPageBy(pageable hateoas.Pageable, criterias map[string]interface{}) (hateoas.Page, error) {
	page := hateoas.NewPage(pageable, defaultPageSize, v1.ContentBasePath)
	var contents []*v1.Content

	if err := repo.db.
		Where(criterias).
		Order(page.Pageable.GetSortClause()).
		Limit(page.Pageable.Size).
		Offset(page.Pageable.GetOffset()).
		Find(&contents).Error; err != nil {
		return page, err
	}
	page.Content = contents

	count := 0
	if err := repo.db.Model(&v1.Content{}).Where(criterias).Count(&count).Error; err != nil {
		return page, err
	}
	page.TotalElements = count

	if pageable.IndexedBy != "" {
		currentIndex := map[string][]*v1.Content{}
		ids := map[string]bool{}
		for _, content := range contents {
			indexedField, err := repo.getIndexedField(pageable.IndexedBy, content)
			if err != nil {
				return page, err
			}
			currentIndex[indexedField] = append(currentIndex[indexedField], content)
			ids[indexedField] = true
		}
		page.Content = currentIndex
		for id := range ids {
			page.IDs = append(page.IDs, id)
		}
	}

	return page, nil
}

func (repo *Repository) getIndexedField(field string, application *v1.Content) (string, error) {
	return "", hateoas.NewUnsupportedIndexError(field)
}

// Save persists an content to the database
func (repo *Repository) Save(content hateoas.Entity) error {
	cont, err := repo.mustBeEntity(content)
	if err != nil {
		return err
	}

	if cont.ID == 0 {
		return repo.db.Create(cont).Error
	}
	return repo.db.Unscoped().Save(cont).Error
}

// Truncate empties the applications table for testing purposes
func (repo *Repository) Truncate() error {
	return repo.db.Delete(v1.Content{}).Error
}

// Remove deletes the application whose GetID is given as a parameter
func (repo *Repository) Remove(app interface{}) error {
	app, err := repo.mustBeEntity(app)
	if err != nil {
		return err
	}

	return repo.db.Delete(app).Error
}

// FindByID gives the details of a particular application
func (repo *Repository) FindByID(id interface{}) (hateoas.Entity, error) {
	app := v1.Content{}
	if err := repo.db.First(&app, id).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

// FindOneByUnscoped gives the details of a particular application, even if soft deleted
func (repo *Repository) FindOneByUnscoped(criterias map[string]interface{}) (hateoas.SoftDeletableEntity, error) {
	app := v1.Content{}
	err := repo.db.Unscoped().Where(criterias).First(&app).Error
	if gorm.IsRecordNotFoundError(err) {
		return &app, hateoas.NewEntityDoesNotExistError(app, criterias)
	}
	return &app, err
}

// FindBy fetch a collection of applications matching each criteria
func (repo *Repository) FindBy(criterias map[string]interface{}) (interface{}, error) {
	var apps []*v1.Content
	err := repo.db.Where(criterias).Find(&apps).Error
	return apps, err
}

// FindOneByDomainNameVersion fetch the first application matching each criteria
func (repo *Repository) FindOneByDomainNameVersion(domain, name, version string) (*v1.Content, error) {
	app := v1.Content{}
	criterias := map[string]interface{}{
		"domain":  domain,
		"name":    name,
		"version": version,
	}
	err := repo.db.First(&app, criterias).Error
	if gorm.IsRecordNotFoundError(err) {
		return &app, hateoas.NewEntityDoesNotExistError(app, criterias)
	}
	return &app, err
}

// FindOneByDomainName fetch the first application matching each criteria
func (repo *Repository) FindOneByDomainName(domain, name string) (*v1.Content, error) {
	app := v1.Content{}
	criterias := map[string]interface{}{
		"domain": domain,
		"name":   name,
	}
	err := repo.db.First(&app, criterias).Error
	if gorm.IsRecordNotFoundError(err) {
		return &app, hateoas.NewEntityDoesNotExistError(app, criterias)
	}
	return &app, err
}

// FindOneBy find by criterias
func (repo *Repository) FindOneBy(criterias map[string]interface{}) (hateoas.Entity, error) {
	app := v1.Content{}
	err := repo.db.Where(criterias).First(&app).Error
	if gorm.IsRecordNotFoundError(err) {
		return &app, hateoas.NewEntityDoesNotExistError(app, criterias)
	}
	return &app, err
}

func (repo *Repository) mustBeEntity(id interface{}) (*v1.Content, error) {
	var app *v1.Content
	switch id := id.(type) {
	case uint:
		app = &v1.Content{ID: id}
	case *v1.Content:
		app = id
	case v1.Content:
		app = &id
	default:
		return nil, hateoas.NewUnsupportedEntityError(app, id)
	}
	return app, nil
}