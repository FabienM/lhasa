package deployment

import (
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"github.com/ovh/lhasa/api/hateoas"
	"github.com/ovh/lhasa/api/v1"
)

const (
	defaultPageSize = 20

	queryUndeployByApplicationNameEnvSlug = "UPDATE \"deployments\" AS \"d\" " +
		"SET \"undeployed_at\" = now(), \"updated_at\" = now() " +
		"FROM \"applications\" as \"a\" " +
		"WHERE \"d\".\"deleted_at\" IS NULL " +
		"AND \"a\".\"id\" = \"d\".\"application_id\" " +
		"AND \"a\".\"domain\" = ? " +
		"AND \"a\".\"name\" = ? " +
		"AND \"d\".\"environment_id\" = ? " +
		"AND \"undeployed_at\" IS NULL"
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

// GetNewEntityInstance returns a new empty instance of the entity managed by this repository
func (repo *Repository) GetNewEntityInstance() hateoas.Entity {
	return &v1.Deployment{}
}

// FindAllPage returns a page of matching entities
func (repo *Repository) FindAllPage(pageable hateoas.Pageable) (hateoas.Page, error) {
	return repo.FindPageBy(pageable, map[string]interface{}{})
}

// FindPageBy returns a page of matching entities
func (repo *Repository) FindPageBy(pageable hateoas.Pageable, criterias map[string]interface{}) (hateoas.Page, error) {
	if pageable.Size == 0 {
		pageable.Size = defaultPageSize
	}
	page := hateoas.Page{Pageable: pageable, BasePath: v1.DeploymentBasePath}
	var deployments []*v1.Deployment

	if err := repo.db.Preload("Environment").Preload("Application").Model(v1.Deployment{}).
		Offset(pageable.Page*pageable.Size).Limit(pageable.Size).Find(&deployments, criterias).Error; err != nil {
		return page, err
	}
	page.Content = deployments

	count := 0
	if err := repo.db.Model(v1.Deployment{}).Where(criterias).Count(&count).Error; err != nil {
		return page, err
	}
	page.TotalElements = count

	return page, nil
}

// FindAll returns all entities of the repository type
func (repo *Repository) FindAll() (interface{}, error) {
	return repo.FindBy(map[string]interface{}{})
}

// Save persists an deployment to the database
func (repo *Repository) Save(deployment hateoas.Entity) error {
	dep, err := repo.mustBeEntity(deployment)
	if err != nil {
		return err
	}

	if dep.ID == 0 {
		publicID, err := uuid.NewV4()
		if err != nil {
			return err
		}
		dep.PublicID = publicID.String()
		return repo.db.Create(dep).Error
	}
	return repo.db.Unscoped().Save(dep).Error
}

// Truncate empties the deployments table for testing purposes
func (repo *Repository) Truncate() error {
	return repo.db.Delete(v1.Deployment{}).Error
}

// Remove deletes the deployment whose GetID is given as a parameter
func (repo *Repository) Remove(dep interface{}) error {
	dep, err := repo.mustBeEntity(dep)
	if err != nil {
		return err
	}

	return repo.db.Delete(dep).Error
}

// FindByID gives the details of a particular deployment
func (repo *Repository) FindByID(id interface{}) (hateoas.Entity, error) {
	dep := v1.Deployment{}
	if err := repo.db.First(&dep, id).Error; err != nil {
		return nil, err
	}
	return &dep, nil
}

// FindOneByUnscoped gives the details of a particular deployment, even if soft deleted
func (repo *Repository) FindOneByUnscoped(criterias map[string]interface{}) (hateoas.SoftDeletableEntity, error) {
	dep := v1.Deployment{}
	err := repo.db.Model(v1.Deployment{}).Unscoped().First(dep, criterias).Error
	if gorm.IsRecordNotFoundError(err) {
		return &dep, hateoas.NewEntityDoesNotExistError(dep, criterias)
	}
	return &dep, err
}

// FindBy fetch a collection of deployments matching each criteria
func (repo *Repository) FindBy(criterias map[string]interface{}) (interface{}, error) {
	var deps []*v1.Deployment
	err := repo.db.Model(v1.Deployment{}).Find(&deps, criterias).Error
	return deps, err
}

// FindActivesBy fetch a collection of deployments matching each criteria
func (repo *Repository) FindActivesBy(criterias map[string]interface{}) ([]*v1.Deployment, error) {
	var deps []*v1.Deployment

	err := repo.db.Preload("Environment").Preload("Application").Model(v1.Deployment{}).Where("undeployed_at IS NULL").Find(&deps, criterias).Error
	return deps, err
}

// FindOneBy fetch the first deployment matching each criteria
func (repo *Repository) FindOneBy(criterias map[string]interface{}) (hateoas.Entity, error) {
	dep := v1.Deployment{}
	err := repo.db.Where(criterias).First(&dep).Error
	if gorm.IsRecordNotFoundError(err) {
		return &dep, hateoas.NewEntityDoesNotExistError(dep, criterias)
	}
	return &dep, err
}

// UndeployByApplicationEnv updates all deployments attached to a given application regardless version with an undeploy date to now
func (repo *Repository) UndeployByApplicationEnv(domain, name string, envID uint) error {
	return repo.db.Exec(queryUndeployByApplicationNameEnvSlug, domain, name, envID).Error
}

func (repo *Repository) mustBeEntity(id interface{}) (*v1.Deployment, error) {
	var dep *v1.Deployment
	switch id := id.(type) {
	case uint:
		dep = &v1.Deployment{ID: id}
	case *v1.Deployment:
		dep = id
	case v1.Deployment:
		dep = &id
	default:
		return nil, hateoas.NewUnsupportedEntityError(dep, id)
	}
	return dep, nil
}