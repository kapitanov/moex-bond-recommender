package data

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Issuer содержит данные эмитента
type Issuer struct {
	ID        int       `gorm:"primaryKey; column:id"`
	MoexID    int       `gorm:"column:moex_id; unique"`
	Name      string    `gorm:"name"`
	INN       *string   `gorm:"inn"`
	OKPO      *string   `gorm:"okpo"`
	CreatedAt time.Time `gorm:"created"`
	UpdatedAt time.Time `gorm:"updated"`
	Bonds     []Bond    `gorm:"foreignKey:IssuerID"`
}

// TableName задает название таблицы
func (Issuer) TableName() string {
	return "issuers"
}

// IssuerRepository отвечает за управление записями в таблице эмитентов
type IssuerRepository interface {
	// GetByID выполняет поиски эмитента по полю Issuer.ID
	// Если эмитент не найден, возвращается ошибка ErrNotFound
	GetByID(id int) (*Issuer, error)

	// GetByMoexID выполняет поиски эмитента по полю Issuer.MoexID
	// Если эмитент не найден, возвращается ошибка ErrNotFound
	GetByMoexID(moexID int) (*Issuer, error)

	// Create создает эмитента
	// Если эмитент уже существует, то возвращается ErrAlreadyExists
	Create(args CreateIssuerArgs) (*Issuer, error)

	// Update обновляет данные эмитента
	// Если эмитент не найден, возвращается ошибка ErrNotFound
	// Если эмитент не найден, возвращается ошибка ErrNotFound
	Update(id int, args UpdateIssuerArgs) (*Issuer, error)
}

// CreateIssuerArgs содержит данные для создания эмитента
type CreateIssuerArgs struct {
	MoexID int
	Name   string
	INN    *string
	OKPO   *string
}

// UpdateIssuerArgs содержит данные для обновления эмитента
type UpdateIssuerArgs struct {
	Name string
	INN  *string
	OKPO *string
}

type issuerRepository struct {
	db *gorm.DB
}

// GetByID выполняет поиски эмитента по полю Issuer.ID
// Если эмитент не найден, возвращается ошибка ErrNotFound
func (repo *issuerRepository) GetByID(id int) (*Issuer, error) {
	var issuer Issuer
	err := repo.db.First(&issuer, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &issuer, nil
}

// GetByMoexID выполняет поиски эмитента по полю Issuer.MoexID
// Если эмитент не найден, возвращается ошибка ErrNotFound
func (repo *issuerRepository) GetByMoexID(moexID int) (*Issuer, error) {
	var issuer Issuer
	err := repo.db.First(&issuer, "moex_id = ?", moexID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &issuer, nil
}

// Create создает эмитента
// Если эмитент уже существует, то возвращается ErrAlreadyExists
func (repo *issuerRepository) Create(args CreateIssuerArgs) (*Issuer, error) {
	_, err := repo.GetByMoexID(args.MoexID)
	if err != nil {
		if err != ErrNotFound {
			return nil, err
		}
	} else {
		return nil, ErrAlreadyExists
	}

	now := time.Now().UTC()
	issuer := Issuer{
		ID:        0,
		MoexID:    args.MoexID,
		Name:      args.Name,
		INN:       args.INN,
		OKPO:      args.OKPO,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = repo.db.Create(&issuer).Error
	if err != nil {
		return nil, err
	}

	return &issuer, nil
}

// Update обновляет данные эмитента
// Если эмитент не найден, возвращается ошибка ErrNotFound
func (repo *issuerRepository) Update(id int, args UpdateIssuerArgs) (*Issuer, error) {
	issuer, err := repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	hasChanges := false
	if issuer.Name != args.Name {
		issuer.Name = args.Name
		hasChanges = true
	}
	if issuer.INN != args.INN {
		issuer.INN = args.INN
		hasChanges = true
	}
	if issuer.OKPO != args.OKPO {
		issuer.OKPO = args.OKPO
		hasChanges = true
	}

	if !hasChanges {
		return issuer, nil
	}

	issuer.UpdatedAt = time.Now().UTC()
	err = repo.db.Updates(issuer).Error
	if err != nil {
		return nil, err
	}

	return issuer, nil
}
