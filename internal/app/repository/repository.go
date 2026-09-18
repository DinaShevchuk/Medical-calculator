package repository

import (
	"Medical_pediatric_calculator/internal/models" // ← правильный путь
	"errors"
	"strconv"
	"strings"
)

type Repository struct {
	services []models.Service
}

func NewRepository() (*Repository, error) {
	services := []models.Service{
		{
			ID:    1,
			Title: "Нурофен",
			Concentration: models.Concentration{
				Mg: 100,
				Ml: 5,
			},
			AdultDose:   10,
			Description: "Противовоспалительное, жаропонижающее средство. Ибупрофен 100 мг/5 мл. Применяется при болях и лихорадке у детей. Не принимать натощак.",
			ImageKey:    "Nurofen.jpg",
			VideoKey:    "Nurofen.mp4",
			Likes:       []int{1, 3, 5, 2, 4, 6, 7, 8, 9, 10, 15},
			Status:      "опубликован",
		},
		{
			ID:    2,
			Title: "Парацетамол",
			Concentration: models.Concentration{
				Mg: 1500,
				Ml: 15,
			},
			AdultDose:   15,
			Description: "Анальгетик и антипиретик. Безопасный жаропонижающий препарат. Максимальная суточная доза 60 мг/кг.",
			ImageKey:    "Paracetamol.jpg",
			VideoKey:    "Paracetamol.mp4",
			Likes:       []int{2, 4},
			Status:      "опубликован",
		},
		{
			ID:    3,
			Title: "Амоксициллин",
			Concentration: models.Concentration{
				Mg: 250,
				Ml: 15,
			},
			AdultDose:   40,
			Description: "Антибиотик пенициллинового ряда. Применяется при бактериальных инфекциях. Курс лечения 5-7 дней.",
			ImageKey:    "Amosikcilin.jpg",
			VideoKey:    "Amosikcilin.mp4",
			Likes:       []int{1},
			Status:      "черновик",
		},
		{
			ID:    4,
			Title: "Активированный уголь",
			Concentration: models.Concentration{
				Mg: 0,
				Ml: 0,
			},
			AdultDose:   5,
			Description: "Энтеросорбент. Применяется при отравлениях и интоксикациях.",
			ImageKey:    "Activated_charcood.jpg",
			VideoKey:    "Activated_charcood.mp4",
			Likes:       []int{},
			Status:      "опубликован",
		},
		{
			ID:    5,
			Title: "Сумамед",
			Concentration: models.Concentration{
				Mg: 0,
				Ml: 0,
			},
			AdultDose:   10,
			Description: "Макролидный антибиотик. Курс лечения 3 дня. Принимается за 1 час до еды.",
			ImageKey:    "Cumamed.jpg",
			VideoKey:    "Cumamed.mp4",
			Likes:       []int{},
			Status:      "удален",
		},
		{
			ID:    6,
			Title: "Аспирин",
			Concentration: models.Concentration{
				Mg: 0,
				Ml: 0,
			},
			AdultDose:   30,
			Description: "Противовоспалительное средство. Ацетилсалициловая кислота. Не рекомендуется детям до 12 лет.",
			ImageKey:    "Aspirin.jpg",
			VideoKey:    "Aspirin.mp4",
			Likes:       []int{},
			Status:      "опубликован",
		},
	}

	return &Repository{services: services}, nil
}

func (r *Repository) GetPublishedServices() []models.Service {
	var result []models.Service
	for _, s := range r.services {
		if s.Status == "опубликован" {
			result = append(result, s)
		}
	}
	return result
}

func (r *Repository) GetServiceByID(id int) (models.Service, error) {
	for _, s := range r.services {
		if s.ID == id {
			if s.Status == "удален" {
				return models.Service{}, errors.New("услуга удалена")
			}
			return s, nil
		}
	}
	return models.Service{}, errors.New("услуга не найдена")
}

func (r *Repository) GetDraftService() (models.Service, error) {
	for _, s := range r.services {
		if s.Status == "черновик" {
			return s, nil
		}
	}
	return models.Service{}, errors.New("черновик не найден")
}

func (r *Repository) GetNextServiceID(currentID int, direction string) (int, error) {
	published := r.GetPublishedServices()
	if len(published) == 0 {
		return 0, errors.New("нет опубликованных услуг")
	}

	currentIndex := -1
	for i, s := range published {
		if s.ID == currentID {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return published[0].ID, nil
	}

	if direction == "true" {
		if currentIndex+1 < len(published) {
			return published[currentIndex+1].ID, nil
		}
		return published[0].ID, nil
	}

	return published[currentIndex].ID, nil
}

func (r *Repository) SearchServices(query string) []models.Service {
	var result []models.Service
	query = strings.TrimSpace(query)

	if query == "" {
		return r.GetPublishedServices()
	}

	dose, err := strconv.Atoi(query)
	if err != nil {
		return result
	}

	for _, s := range r.GetPublishedServices() {
		if s.AdultDose == dose {
			result = append(result, s)
		}
	}

	return result
}
