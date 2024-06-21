package handlers

import (
	"github.com/rs/zerolog/log"
	"net/http"

	"github.com/babylonchain/staking-api-service/internal/types"
	"github.com/babylonchain/staking-api-service/internal/utils"
)

// GetStakerDelegations @Summary Get staker delegations
// @Description Retrieves delegations for a given staker
// @Produce json
// @Param staker_btc_pk query string true "Staker BTC Public Key"
// @Param pagination_key query string false "Pagination key to fetch the next page of delegations"
// @Success 200 {object} PublicResponse[[]services.DelegationPublic]{array} "List of delegations and pagination token"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v1/staker/delegations [get]
func (h *Handler) GetStakerDelegations(request *http.Request) (*Result, *types.Error) {
	stakerBtcPk, err := parsePublicKeyQuery(request, "staker_btc_pk")
	if err != nil {
		return nil, err
	}
	paginationKey, err := parsePaginationQuery(request)
	if err != nil {
		return nil, err
	}

	delegations, newPaginationKey, err := h.services.DelegationsByStakerPk(request.Context(), stakerBtcPk, paginationKey)
	if err != nil {
		return nil, err
	}

	return NewResultWithPagination(delegations, newPaginationKey), nil
}

func (h *Handler) GetStakerCountByFinalityProvider(request *http.Request) (*Result, *types.Error) {
	finalityProviderPkHex, err := parsePublicKeyQuery(request, "finality_provider_pk_hex")
	if err != nil {
		return nil, err
	}

	log.Ctx(request.Context()).Debug().Msgf("GetStakerCountByFinalityProvider: finalityProviderPkHex:%s",
		finalityProviderPkHex)

	delegations, err := h.services.DelegationsByFinalityProviderPkHex(request.Context(), finalityProviderPkHex)
	if err != nil {
		return nil, err
	}

	log.Ctx(request.Context()).Debug().Msgf("GetStakerCountByFinalityProvider: delegations len:%d",
		len(delegations))

	countMap := make(map[string]int)
	for _, delegation := range delegations {
		if delegation.State == types.Unbonded.ToString() {
			continue
		}
		countMap[delegation.StakerPkHex]++
	}

	return NewResult(len(countMap)), nil
}

func (h *Handler) GetDelegationsCountByFinalityProvider(request *http.Request) (*Result, *types.Error) {
	finalityProviderPkHex, err := parsePublicKeyQuery(request, "finality_provider_pk_hex")
	if err != nil {
		return nil, err
	}

	log.Ctx(request.Context()).Debug().Msgf("GetDelegationsCountByFinalityProvider-: finalityProviderPkHex:%s",
		finalityProviderPkHex)

	delegations, err := h.services.DelegationsByFinalityProviderPkHex(request.Context(), finalityProviderPkHex)
	if err != nil {
		return nil, err
	}

	log.Ctx(request.Context()).Debug().Msgf("GetDelegationsCountByFinalityProvider: delegations len:%d",
		len(delegations))

	countSlice := make([]string, 0)
	for _, delegation := range delegations {
		if delegation.State == types.Unbonded.ToString() {
			continue
		}
		countSlice = append(countSlice, delegation.StakerPkHex)
	}

	return NewResult(len(countSlice)), nil
}

func (h *Handler) GetStakerCountByStakerPk(request *http.Request) (*Result, *types.Error) {
	finalityProviderPkHex, err := parsePublicKeyQuery(request, "finality_provider_pk_hex")
	if err != nil {
		return nil, err
	}

	log.Ctx(request.Context()).Debug().Msgf("GetStakerCountByStakerPk: finalityProviderPkHex:%s", finalityProviderPkHex)

	count, err := h.services.StakerCountByStakerPk(request.Context(), finalityProviderPkHex)
	if err != nil {
		return nil, err
	}

	return NewResult(count), nil
}

// CheckStakerDelegationExist @Summary Check if a staker has an active delegation
// @Description Check if a staker has an active delegation by the staker BTC address (Taproot only)
// @Description Optionally, you can provide a timeframe to check if the delegation is active within the provided timeframe
// @Description The available timeframe is "today" which checks after UTC 12AM of the current day
// @Produce json
// @Param address query string true "Staker BTC address in Taproot format"
// @Param timeframe query string false "Check if the delegation is active within the provided timeframe" Enums(today)
// @Success 200 {object} Result "Result"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v1/staker/delegation/check [get]
func (h *Handler) CheckStakerDelegationExist(request *http.Request) (*Result, *types.Error) {
	address, err := parseBtcAddressQuery(request, "address", h.config.Server.BTCNetParam)
	if err != nil {
		return nil, err
	}

	afterTimestamp, err := parseTimeframeToAfterTimestamp(request.URL.Query().Get("timeframe"))
	if err != nil {
		return nil, err
	}

	exist, err := h.services.CheckStakerHasActiveDelegationByAddress(
		request.Context(), address, afterTimestamp,
	)
	if err != nil {
		return nil, err
	}

	return NewResult(exist), nil
}

func parseTimeframeToAfterTimestamp(timeframe string) (int64, *types.Error) {
	switch timeframe {
	case "": // We ignore and return 0 if no timeframe is provided
		return 0, nil
	case "today":
		return utils.GetTodayStartTimestampInSeconds(), nil
	default:
		return 0, types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, "invalid timeframe value",
		)
	}
}
