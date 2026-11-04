package repository

import (
	"dosage_drugs/internal/app/ds"
)

// RegisterUser сохраняет пользователя. В password уже должен лежать bcrypt-хеш.
func (r *Repository) RegisterUser(login, passwordHash string) (*ds.User, error) {
	u := &ds.User{Login: login, Password: passwordHash}
	if err := r.db.Create(u).Error; err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repository) GetUserByLogin(login string) (*ds.User, error) {
	var u ds.User
	if err := r.db.Where("login = ?", login).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetUserByID(id int) (*ds.User, error) {
	var u ds.User
	if err := r.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
