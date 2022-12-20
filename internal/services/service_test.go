package services

import (
	"XTechProject/internal/models"
	mockServices "XTechProject/internal/services/mocks"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestGetResponse(t *testing.T) {
	link := "https://github.com/AlexanderValov"
	_, err := getResponse(link)
	require.NoError(t, err)
}

func TestGetResponseError(t *testing.T) {
	link := "https:/link.csa"
	_, err := getResponse(link)
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
	srv := mockServices.NewMockServicer(ctl)
	unixTime := int64(1671542754)
	lastValue := "666.6"
	srv.EXPECT().UpdateBTCInDB(unixTime, lastValue)
	err := srv.UpdateBTCInDB(unixTime, lastValue)
	require.NoError(t, err)
}
