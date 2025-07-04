package service

import (
	"context"
	"github.com/carinfinin/keeper/internal/clientcfg"
	"github.com/carinfinin/keeper/internal/keystore"
	"github.com/carinfinin/keeper/internal/store/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestClientService(t *testing.T) {
	type tt struct {
		name  string
		item  models.Item
		error bool
	}
	tests := []tt{
		{
			name: "positive",
			item: models.Item{UID: "234298", Type: "text", Data: []byte("hello world")},
		},
		{
			name: "positive",
			item: models.Item{UID: "2342899", Type: "text", Data: []byte("hello world8")},
		},
	}

	cfg, err := clientcfg.LoadConfig()
	require.NoError(t, err)
	s, err := NewClientService(cfg)
	require.NoError(t, err)
	defer s.Close()

	err = keystore.SaveDerivedKey("test", "test")
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := test.item.Data
			err = s.AddDecryptedItem(context.Background(), &test.item, test.item.Data)
			assert.NoError(t, err)

			item, err := s.GetDecryptedItem(context.Background(), test.item.UID)
			assert.NoError(t, err)

			assert.Equal(t, item.Data, data)

			err = s.DeleteItem(context.Background(), test.item.UID)
			assert.NoError(t, err)

			_, err = s.GetDecryptedItem(context.Background(), test.item.UID)
			assert.Error(t, err)

		})
	}
}
