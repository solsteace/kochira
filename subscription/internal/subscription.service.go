package internal

import (
	"slices"
	"time"

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

func (ss SubscriptionService) Init(userId []uint64) error {
	existingSubscriptions, err := ss.repo.CheckManyByOwner(userId)
	if err != nil {
		return err
	}

	now := time.Now()
	subscriptions := []domain.Subscription{}
	for _, uId := range userId {
		// Number of new users within small time window (say, 2 - 3 seconds) won't
		// be large enough, so using linear search would be okay for now
		if slices.Contains(existingSubscriptions, uId) {
			continue
		}

		s, err := domain.NewSubscription(nil, uId, now)
		if err != nil {
			return err
		}
		subscriptions = append(subscriptions, s)
	}

	return ss.repo.Create(subscriptions)
}

func (ss SubscriptionService) Order(
	userId uint64,
	packageName string,
	quantity uint,
) {

}

func (ss SubscriptionService) Extend(userId uint64, packageName string) {
}
