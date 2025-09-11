package internal

import (
	"github.com/solsteace/kochira/subscription/internal/domain"
	"github.com/solsteace/kochira/subscription/internal/repository"
)

type SubscriptionService struct {
	repo repository.Subscription
}

func NewSubscriptionService(repo repository.Subscription) SubscriptionService {
	return SubscriptionService{repo}
}

func (ss SubscriptionService) GetByUserId(id uint64) (domain.Subscription, error) {
	return ss.repo.GetByOwner(id)
}

func (ss SubscriptionService) InferPerks(userId uint64) {

}

func (ss SubscriptionService) Order(
	userId uint64,
	packageName string,
	quantity uint,
) {

}

func (ss SubscriptionService) Extend(userId uint64, packageName string) {
}
