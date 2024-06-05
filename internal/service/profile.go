package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/sirupsen/logrus"
)

type ProfileRepositoryPSQL interface {
	GetProfile(ctx context.Context, userId int) (*core.User, error)
	UpdateProfile(ctx context.Context, user *core.User) error
	SaveAvatar(ctx context.Context, avatar *core.UserAvatar) error
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

func (p *Profile) GetProfile(ctx context.Context, userId int) (*core.GetProfileResp, error) {
	user, err := p.psqlRepo.GetProfile(ctx, userId)
	if err != nil {
		return nil, err
	}

	userAvatars, err := p.psqlRepo.GetAvatars(ctx, user.ID)
	if err != nil {
		if errors.Is(err, core.ErrAvatarNotFound) {
			return &core.GetProfileResp{
				ID:       user.ID,
				Username: user.Username,
				Phone:    user.Phone,
			}, nil
		}

		return nil, err
	}

	key := fmt.Sprintf("%s_%d_%d.jpg", p.avatarKeySalt, user.ID, userAvatars[len(userAvatars)-1].ID)

	resp, err := p.s3Repo.GetAvatars(ctx, key)
	if err != nil {
		return nil, err
	}

	return &core.GetProfileResp{
		ID:        user.ID,
		Username:  user.Username,
		Phone:     user.Phone,
		AvatarUrl: resp.URL,
	}, nil
}

func (p *Profile) UpdateProfile(ctx context.Context, user *core.User) error {
	if err := p.psqlRepo.UpdateProfile(ctx, user); err != nil {
		return core.ErrThisCredIsAlready
	}

	return nil
}

func (p *Profile) UploadAvatar(ctx context.Context, file multipart.File, userId int) error {
	avatar := &core.UserAvatar{
		UserID: userId,
	}

	if err := p.psqlRepo.SaveAvatar(ctx, avatar); err != nil {
		return err
	}

	key := fmt.Sprintf("%s_%d_%d.jpg", p.avatarKeySalt, userId, avatar.ID)

	if err := p.s3Repo.UploadAvatar(ctx, file, key); err != nil {
		return err
	}

	return nil
}

func (p *Profile) GetAvatars(ctx context.Context, userId int) ([]*core.GetAllUserAvatarsResp, error) {
	avatars, err := p.psqlRepo.GetAvatars(ctx, userId)
	if err != nil {
		if errors.Is(err, core.ErrAvatarNotFound) {
			return nil, nil
		}

		return nil, err
	}

	var response []*core.GetAllUserAvatarsResp
	for _, avatar := range avatars {
		key := fmt.Sprintf("%s_%d_%d.jpg", p.avatarKeySalt, userId, avatar.ID)

		output, err := p.s3Repo.GetAvatars(ctx, key)
		if err != nil {
			return nil, err
		}

		response = append(response, &core.GetAllUserAvatarsResp{
			ID:        avatar.ID,
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
