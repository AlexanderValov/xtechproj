package services

import (
	"XTechProject/cmd/config"
	"XTechProject/internal/models"
	mock_repository "XTechProject/internal/repository/mocks"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
	"time"
)

func TestGetResponse(t *testing.T) {
	link := "https://github.com/AlexanderValov"
	_, err := getResponse(link)
	require.NoError(t, err)
}

func TestGetResponseError(t *testing.T) {
	link := "https://github.com/AlexanderValov12344"
	_, err := getResponse(link)
	require.Error(t, err)
	link = "https://13.com/2"
	_, err = getResponse(link)
	require.Error(t, err)
}

func TestSerializeFiatCurrenciesData(t *testing.T) {
	// test with 101 instance (check that response will have the same len)
	correctInput := []Valute{
		{
			ID:       "4",
			NumCode:  "4",
			CharCode: "USD",
			Nominal:  "1",
			Name:     "USDValute",
			Value:    "80,551",
		},
	}
	for i := 0; i < 100; i++ {
		correctInput = append(correctInput, Valute{
			ID:       "test",
			NumCode:  "test",
			CharCode: "test",
			Nominal:  "12",
			Name:     "test",
			Value:    "123,4",
		})
	}
	data, usdrub, err := serializeFiatCurrenciesData(correctInput)
	require.NoError(t, err)
	var expData []models.Currency
	_ = json.Unmarshal(data, &expData)
	require.Equal(t, len(expData), len(correctInput))
	require.Equal(t, usdrub, 80.551)
	// test to check that struct generate ok
	correctInput = []Valute{
		{
			ID:       "4",
			NumCode:  "4",
			CharCode: "USD",
			Nominal:  "1",
			Name:     "USDValute",
			Value:    "80,551",
		},
	}
	cur := []models.Currency{
		{
			ID:       correctInput[0].ID,
			Name:     correctInput[0].Name,
			Nominal:  1,
			CharCode: correctInput[0].CharCode,
			NumCode:  correctInput[0].NumCode,
			Val:      80.551,
		},
	}
	data, _, err = serializeFiatCurrenciesData(correctInput)
	require.NoError(t, err)
	marshal, _ := json.Marshal(cur)
	require.Equal(t, marshal, data)

}

func TestSerializeFiatCurrenciesDataError(t *testing.T) {
	cases := []struct {
		name   string
		input  []Valute
		expErr error
	}{
		{
			name: "case without USD",
			input: []Valute{
				{
					ID:       "1",
					NumCode:  "1",
					CharCode: "FIRST",
					Nominal:  "10",
					Name:     "FirstValute",
					Value:    "11,0011",
				},
			},
			expErr: ErrUSDNotFound,
		}, {
			name:   "case with empty []Valute",
			input:  []Valute{},
			expErr: ErrEmptyValuteSlice,
		},
		{
			name: "case with wrong Valute.Value",
			input: []Valute{
				{
					ID:       "4",
					NumCode:  "4",
					CharCode: "USD",
					Nominal:  "1",
					Name:     "USDValute",
					Value:    "wrongValue",
				},
			},
			expErr: strconv.ErrSyntax,
		},
		{
			name: "case with wrong Valute.Nominal",
			input: []Valute{
				{
					ID:       "4",
					NumCode:  "4",
					CharCode: "USD",
					Nominal:  "wrongNominal",
					Name:     "USDValute",
					Value:    "123,1",
				},
			},
			expErr: strconv.ErrSyntax,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, _, err := serializeFiatCurrenciesData(c.input)
			require.ErrorIs(t, err, c.expErr)
		})
	}
}

func TestSerializeOrderBy(t *testing.T) {
	cases := []struct {
		name  string
		input string
		exp   string
	}{
		{
			name:  "test with value",
			input: "value",
			exp:   "ORDER BY value",
		},
		{
			name:  "test with -value",
			input: "-value",
			exp:   "ORDER BY value DESC",
		},
		{
			name:  "test with created_at",
			input: "created_at",
			exp:   "ORDER BY created_at",
		},
		{
			name:  "test with -created_at",
			input: "-created_at",
			exp:   "ORDER BY created_at DESC",
		},
		{
			name:  "test with latest",
			input: "latest",
			exp:   "ORDER BY latest",
		},
		{
			name:  "test with -latest",
			input: "-latest",
			exp:   "ORDER BY latest DESC",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ans, err := serializeOrderBy(c.input)
			require.NoError(t, err)
			require.Equal(t, c.exp, ans)
		})
	}
}

func TestSerializeOrderByError(t *testing.T) {
	ans, err := serializeOrderBy("wrong_order_by")
	require.Error(t, err)
	require.Equal(t, "", ans)
}

func TestUpdateBTCInDB(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	repo := mock_repository.NewMockRepositorier(ctl)
	cfg, err := config.New()
	if err != nil {
		require.NoError(t, err)
	}
	tm := time.Now()
	cur := []models.Currency{
		{
			ID:       "1",
			Name:     "test",
			Nominal:  1,
			CharCode: "test",
			NumCode:  "123",
			Val:      80.551,
		},
		{
			ID:       "2",
			Name:     "test1",
			Nominal:  1,
			CharCode: "USD",
			NumCode:  "1234",
			Val:      70.551,
		},
	}
	expFiat := &models.Fiat{
		ID:        1,
		Latest:    true,
		CreatedAt: &tm,
		USDRUB:    cur[1].Val,
	}
	expFiat.Currencies, err = json.Marshal(cur)
	if err != nil {
		require.NoError(t, err)
	}
	unixTime := int64(1671542754)
	lastValue := "666.6"
	btc := &models.BTC{
		ID:        0,
		Value:     666.6,
		Latest:    true,
		CreatedAt: unixTimeToTime(unixTime),
	}
	btcToFiat, err := calculateBTCToFiat(cur, btc.Value, cur[1].Val)
	if err != nil {
		require.NoError(t, err)
	}
	btc.BTCToFiat, err = json.Marshal(btcToFiat)
	if err != nil {
		require.NoError(t, err)
	}
	repo.EXPECT().UpdateLastRecordForBTC().Return(nil).Times(1)
	repo.EXPECT().GetLastFiat().Return(expFiat, nil).Times(1)
	repo.EXPECT().CreateBTCRecord(btc).Return(nil).Times(1)
	srv := NewManagementService(repo, cfg)
	err = srv.UpdateBTCInDB(unixTime, lastValue)
	require.NoError(t, err)
}
