package service

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/sirupsen/logrus"
)

type ProfileRepositoryPSQL interface {
	GetProfile(ctx context.Context, userId int) (*core.User, error)
	UpdateProfile(ctx context.Context, user *core.User) (*core.User, error)
	UploadAvatar(ctx context.Context, avatar *core.UserAvatar) (*core.UserAvatar, error)
	GetAvatars(ctx context.Context, userId int) ([]*core.UserAvatar, error)
	GetAvatar(ctx context.Context, avatarId int) (*core.UserAvatar, error)
	DeleteAvatar(ctx context.Context, id int) error
}

type ProfileRepositoryS3 interface {
	GetAvatars(ctx context.Context, key string) (*v4.PresignedHTTPRequest, error)
	UploadAvatar(ctx context.Context, file multipart.File, key string) error
	DeleteAvatar(ctx context.Context, key string) error
}

type Profile struct {
	psqlRepo ProfileRepositoryPSQL
	s3Repo   ProfileRepositoryS3

	avatarKeySalt string

	log *logrus.Logger
}

func NewProfile(psqlRepo ProfileRepositoryPSQL, s3Repo ProfileRepositoryS3, avatarKeySalt string, log *logrus.Logger) *Profile {
	return &Profile{
		psqlRepo: psqlRepo,
		s3Repo:   s3Repo,

		avatarKeySalt: avatarKeySalt,

		log: log,
	}
}

func (p *Profile) GetProfile(ctx context.Context, userId int) (*core.GetProfileResponse, error) {
	user, err := p.psqlRepo.GetProfile(ctx, userId)
	if err != nil {
		return nil, err
	}

	userAvatars, err := p.psqlRepo.GetAvatars(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("%s_%d_%d.jpg", p.avatarKeySalt, user.ID, getLastAvatarId(userAvatars))

	resp, err := p.s3Repo.GetAvatars(ctx, key)
	if err != nil {
		return nil, err
	}

	return &core.GetProfileResponse{
		ID:        user.ID,
		Username:  user.Username,
		Phone:     user.Phone,
		AvatarUrl: resp.URL,
	}, nil
}

func getLastAvatarId(avatars []*core.UserAvatar) int {
	if len(avatars) == 0 {
		return 0
	}

	return avatars[len(avatars)-1].ID
}

func (p *Profile) UpdateProfile(ctx context.Context, user *core.User) (*core.User, error) {
	user, err := p.psqlRepo.UpdateProfile(ctx, user)
	if err != nil {
		return nil, core.ErrThisCredIsAlready
	}

	return user, nil
}

func (p *Profile) UploadAvatar(ctx context.Context, file multipart.File, id int) error {
	avatar := &core.UserAvatar{
		UserId: id,
	}

	avatarWithId, err := p.psqlRepo.UploadAvatar(ctx, avatar)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s_%d_%d.jpg", p.avatarKeySalt, id, avatarWithId.ID)

	if err := p.s3Repo.UploadAvatar(ctx, file, key); err != nil {
		return err
	}

	return nil
}

func (p *Profile) GetAvatars(ctx context.Context, userId int) ([]*core.GetAllUSerAvatarsResponse, error) {
	userAvatars, err := p.psqlRepo.GetAvatars(ctx, userId)
	if err != nil {
		return nil, err
	}

	var response []*core.GetAllUSerAvatarsResponse

	for _, avatar := range userAvatars {
		key := fmt.Sprintf("%s_%d_%d.jpg", p.avatarKeySalt, userId, avatar.ID)

		output, err := p.s3Repo.GetAvatars(ctx, key)
		if err != nil {
			return nil, err
		}

		response = append(response, &core.GetAllUSerAvatarsResponse{
			ID:        avatar.ID,
			UserId:    avatar.UserId,
			AvatarUrl: output.URL,
		})
	}

	return response, nil
}

func (p *Profile) DeleteAvatar(ctx context.Context, userId int, avatarId int) error {
	if _, err := p.psqlRepo.GetAvatar(ctx, avatarId); err != nil {
		return core.ErrAvatarNotFound
	}

	key := fmt.Sprintf("%s_%d_%d.jpg", p.avatarKeySalt, userId, avatarId)

	if err := p.s3Repo.DeleteAvatar(ctx, key); err != nil {
		return err
	}

	if err := p.psqlRepo.DeleteAvatar(ctx, avatarId); err != nil {
		return err
	}

	return nil
}
