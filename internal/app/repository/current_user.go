package repository

import "dosage_drugs/internal/app/ds"

var currentUser *ds.User

// CurrentUser — singleton, возвращает зафиксированного пользователя.
// В 4-й ЛР заменится на реальную авторизацию.
func CurrentUser() *ds.User {
    if currentUser == nil {
        currentUser = &ds.User{
            ID:          1,
            Login:       "test",
            Password:    "test123",
            IsModerator: false,
        }
    }
    return currentUser
}
