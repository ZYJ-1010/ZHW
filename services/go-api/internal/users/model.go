package users

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidProfile   = errors.New("invalid user profile")
	ErrPhoneAlreadyUsed = errors.New("phone already used")
)

type User struct {
	ID             int64     `json:"id"`
	OpenID         string    `json:"openId"`
	PhoneMasked    string    `json:"phoneMasked,omitempty"`
	Nickname       string    `json:"nickname"`
	AvatarURL      string    `json:"avatarUrl"`
	AvatarFileID   int64     `json:"avatarFileId,omitempty"`
	RealnameStatus string    `json:"realnameStatus"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
}

type Filter struct {
	Status         string
	RealnameStatus string
	Keyword        string
}

type Store struct {
	nextID  int64
	byID    map[int64]User
	byOpen  map[string]int64
	byPhone map[string]int64
	repo    Repository
}

func NewStore() *Store {
	return NewStoreWithRepository(nil)
}

func NewStoreWithRepository(repo Repository) *Store {
	return &Store{
		nextID:  1,
		byID:    make(map[int64]User),
		byOpen:  make(map[string]int64),
		byPhone: make(map[string]int64),
		repo:    repo,
	}
}

type Repository interface {
	FindByOpenID(ctx context.Context, openID string) (User, bool, error)
	FindByPhoneHash(ctx context.Context, phoneHash string) (User, bool, error)
	FindByID(ctx context.Context, id int64) (User, bool, error)
	List(ctx context.Context, filter Filter) ([]User, error)
	CreateWithOpenID(ctx context.Context, openID string) (User, error)
	CreateWithPhone(ctx context.Context, phoneHash string, phoneMasked string) (User, error)
	UpdatePhoneAuth(ctx context.Context, userID int64, phoneHash string, phoneMasked string) (User, error)
	UpdateProfile(ctx context.Context, userID int64, nickname string, avatarURL string, avatarFileID int64) (User, error)
	UpdateRealnameStatus(ctx context.Context, userID int64, status string) (User, error)
}

func (s *Store) FindByPhoneHash(phoneHash string) (User, bool, error) {
	phoneHash = strings.TrimSpace(phoneHash)
	if phoneHash == "" {
		return User{}, false, nil
	}
	if s.repo != nil {
		user, ok, err := s.repo.FindByPhoneHash(context.Background(), phoneHash)
		if err != nil || ok {
			if ok {
				s.byID[user.ID] = user
				s.byPhone[phoneHash] = user.ID
			}
			return user, ok, err
		}
	}
	id, ok := s.byPhone[phoneHash]
	if !ok {
		return User{}, false, nil
	}
	return s.byID[id], true, nil
}

func (s *Store) FindByOpenID(openID string) (User, bool, error) {
	if s.repo != nil {
		user, ok, err := s.repo.FindByOpenID(context.Background(), openID)
		if err != nil || ok {
			return user, ok, err
		}
	}
	id, ok := s.byOpen[openID]
	if !ok {
		return User{}, false, nil
	}
	return s.byID[id], true, nil
}

func (s *Store) FindByID(id int64) (User, bool, error) {
	if s.repo != nil {
		user, ok, err := s.repo.FindByID(context.Background(), id)
		if err != nil || ok {
			return user, ok, err
		}
	}
	user, ok := s.byID[id]
	return user, ok, nil
}

func (s *Store) List(filter Filter) ([]User, error) {
	filter = normalizeFilter(filter)
	if s.repo != nil {
		return s.repo.List(context.Background(), filter)
	}
	items := make([]User, 0, len(s.byID))
	for _, user := range s.byID {
		if filter.Status != "" && user.Status != filter.Status {
			continue
		}
		if filter.RealnameStatus != "" && user.RealnameStatus != filter.RealnameStatus {
			continue
		}
		if filter.Keyword != "" && !userMatchesKeyword(user, filter.Keyword) {
			continue
		}
		items = append(items, user)
	}
	sort.Slice(items, func(i, j int) bool {
		if !items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].CreatedAt.After(items[j].CreatedAt)
		}
		return items[i].ID > items[j].ID
	})
	return items, nil
}

func (s *Store) Create(openID string) (User, error) {
	if s.repo != nil {
		user, err := s.repo.CreateWithOpenID(context.Background(), openID)
		if err != nil {
			return User{}, err
		}
		s.byID[user.ID] = user
		s.byOpen[openID] = user.ID
		return user, nil
	}
	user := User{
		ID:             s.nextID,
		OpenID:         openID,
		RealnameStatus: "pending",
		Status:         "active",
		CreatedAt:      time.Now(),
	}
	s.nextID++
	s.byID[user.ID] = user
	s.byOpen[openID] = user.ID
	return user, nil
}

func (s *Store) CreateWithPhone(phoneHash string, phoneMasked string) (User, error) {
	phoneHash = strings.TrimSpace(phoneHash)
	phoneMasked = strings.TrimSpace(phoneMasked)
	if phoneHash == "" || phoneMasked == "" {
		return User{}, ErrInvalidProfile
	}
	if s.repo != nil {
		user, err := s.repo.CreateWithPhone(context.Background(), phoneHash, phoneMasked)
		if err != nil {
			return User{}, err
		}
		s.byID[user.ID] = user
		s.byPhone[phoneHash] = user.ID
		return user, nil
	}
	if id, ok := s.byPhone[phoneHash]; ok {
		return s.byID[id], nil
	}
	user := User{
		ID:             s.nextID,
		PhoneMasked:    phoneMasked,
		RealnameStatus: "pending",
		Status:         "active",
		CreatedAt:      time.Now(),
	}
	s.nextID++
	s.byID[user.ID] = user
	s.byPhone[phoneHash] = user.ID
	return user, nil
}

func (s *Store) BindPhoneAuth(userID int64, phoneHash string, phoneMasked string) (User, error) {
	phoneHash = strings.TrimSpace(phoneHash)
	phoneMasked = strings.TrimSpace(phoneMasked)
	if userID <= 0 || phoneHash == "" || phoneMasked == "" {
		return User{}, ErrInvalidProfile
	}
	if existingID, ok := s.byPhone[phoneHash]; ok && existingID != userID {
		return User{}, ErrPhoneAlreadyUsed
	}
	if s.repo != nil {
		if existing, ok, err := s.repo.FindByPhoneHash(context.Background(), phoneHash); err != nil {
			return User{}, err
		} else if ok && existing.ID != userID {
			return User{}, ErrPhoneAlreadyUsed
		}
		user, err := s.repo.UpdatePhoneAuth(context.Background(), userID, phoneHash, phoneMasked)
		if err != nil {
			return User{}, err
		}
		s.byID[user.ID] = user
		s.byPhone[phoneHash] = user.ID
		return user, nil
	}
	user, ok := s.byID[userID]
	if !ok {
		return User{}, ErrInvalidProfile
	}
	user.PhoneMasked = phoneMasked
	s.byID[user.ID] = user
	s.byPhone[phoneHash] = user.ID
	return user, nil
}

func normalizeFilter(filter Filter) Filter {
	filter.Status = strings.TrimSpace(filter.Status)
	filter.RealnameStatus = strings.TrimSpace(filter.RealnameStatus)
	filter.Keyword = strings.ToLower(strings.TrimSpace(filter.Keyword))
	return filter
}

func userMatchesKeyword(user User, keyword string) bool {
	if keyword == "" {
		return true
	}
	return strings.Contains(strings.ToLower(user.Nickname), keyword) ||
		strings.Contains(strings.ToLower(user.OpenID), keyword) ||
		strings.Contains(strings.ToLower(user.PhoneMasked), keyword) ||
		strings.Contains(strconv.FormatInt(user.ID, 10), keyword)
}

func (s *Store) UpdateProfile(userID int64, nickname string, avatarURL string, avatarFileID int64) (User, error) {
	nickname = strings.TrimSpace(nickname)
	avatarURL = strings.TrimSpace(avatarURL)
	if len([]rune(nickname)) > 32 || len(avatarURL) > 500 || avatarFileID < 0 {
		return User{}, ErrInvalidProfile
	}
	if s.repo != nil {
		user, err := s.repo.UpdateProfile(context.Background(), userID, nickname, avatarURL, avatarFileID)
		if err != nil {
			return User{}, err
		}
		s.byID[user.ID] = user
		if user.OpenID != "" {
			s.byOpen[user.OpenID] = user.ID
		}
		return user, nil
	}
	user, ok := s.byID[userID]
	if !ok {
		return User{}, ErrInvalidProfile
	}
	user.Nickname = nickname
	user.AvatarURL = avatarURL
	user.AvatarFileID = avatarFileID
	s.byID[user.ID] = user
	if user.OpenID != "" {
		s.byOpen[user.OpenID] = user.ID
	}
	return user, nil
}

func (s *Store) UpdateRealnameStatus(userID int64, status string) (User, error) {
	status = strings.TrimSpace(status)
	if userID <= 0 || status == "" || len(status) > 32 {
		return User{}, ErrInvalidProfile
	}
	if s.repo != nil {
		user, err := s.repo.UpdateRealnameStatus(context.Background(), userID, status)
		if err != nil {
			return User{}, err
		}
		s.byID[user.ID] = user
		if user.OpenID != "" {
			s.byOpen[user.OpenID] = user.ID
		}
		return user, nil
	}
	user, ok := s.byID[userID]
	if !ok {
		return User{}, ErrInvalidProfile
	}
	user.RealnameStatus = status
	s.byID[user.ID] = user
	return user, nil
}
