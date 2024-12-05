package service

import (
	"context"

	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

type nodeService struct {
}

func (d *nodeService) Drain(ctx context.Context, name string) error {
	err := kom.DefaultCluster().WithContext(ctx).
		Resource(&v1.Node{}).Name(name).
		Ctl().Node().Drain()
	if err != nil {
		return err
	}
	return nil
}
func (d *nodeService) Cordon(ctx context.Context, name string) error {
	err := kom.DefaultCluster().WithContext(ctx).
		Resource(&v1.Node{}).Name(name).
		Ctl().Node().Cordon()
	if err != nil {
		return err
	}
	return nil
}
func (d *nodeService) UnCordon(ctx context.Context, name string) error {
	err := kom.DefaultCluster().WithContext(ctx).
		Resource(&v1.Node{}).Name(name).
		Ctl().Node().UnCordon()
	if err != nil {
		return err
	}
	return nil
}
