// SPDX-FileCopyrightText: Copyright 2024 The Minder Authors
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.uber.org/mock/gomock"
	mock_v1 "github.com/mindersec/minder/pkg/datasources/v1/mock"
)

func TestDataSourceRegistry_RegisterDataSource(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFuncDef := mock_v1.NewMockDataSourceFuncDef(ctrl)
	
	tests := []struct {
		name       string
		dsName     string
		funcs      map[DataSourceFuncKey]DataSourceFuncDef
		wantErr    bool
		errorIs    error
	}{
		{
			name:   "register successful",
			dsName: "testds",
			funcs: map[DataSourceFuncKey]DataSourceFuncDef{
				DataSourceFuncKey("func1"): mockFuncDef,
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reg := NewDataSourceRegistry()
			mockDS := mock_v1.NewMockDataSource(ctrl)
			mockDS.EXPECT().GetFuncs().Return(tc.funcs).AnyTimes()

			err := reg.RegisterDataSource(tc.dsName, mockDS)
			if tc.wantErr {
				require.Error(t, err)
				if tc.errorIs != nil {
					require.ErrorIs(t, err, tc.errorIs)
				}
			} else {
				require.NoError(t, err)

				funcs := reg.GetFuncs()
				require.Len(t, funcs, len(tc.funcs))

				for k, f := range tc.funcs {
					expectedKey := makeKey(tc.dsName, k)
					require.Equal(t, f, funcs[expectedKey])
				}
			}
		})
	}
}

func TestDataSourceRegistry_DuplicateRegistration(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFuncDef := mock_v1.NewMockDataSourceFuncDef(ctrl)

	reg := NewDataSourceRegistry()
	mockDS := mock_v1.NewMockDataSource(ctrl)
	mockDS.EXPECT().GetFuncs().Return(map[DataSourceFuncKey]DataSourceFuncDef{
		DataSourceFuncKey("func1"): mockFuncDef,
	}).AnyTimes()

	err := reg.RegisterDataSource("testds", mockDS)
	require.NoError(t, err)

	// Registering again with same name and same function key should fail
	err = reg.RegisterDataSource("testds", mockDS)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrDuplicateDataSourceFuncKey)
}

func TestDataSourceRegistry_GetFuncs(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFuncDef1 := mock_v1.NewMockDataSourceFuncDef(ctrl)
	mockFuncDef2 := mock_v1.NewMockDataSourceFuncDef(ctrl)

	reg := NewDataSourceRegistry()
	mockDS1 := mock_v1.NewMockDataSource(ctrl)
	mockDS1.EXPECT().GetFuncs().Return(map[DataSourceFuncKey]DataSourceFuncDef{
		DataSourceFuncKey("func1"): mockFuncDef1,
	}).AnyTimes()

	mockDS2 := mock_v1.NewMockDataSource(ctrl)
	mockDS2.EXPECT().GetFuncs().Return(map[DataSourceFuncKey]DataSourceFuncDef{
		DataSourceFuncKey("func2"): mockFuncDef2,
	}).AnyTimes()

	require.NoError(t, reg.RegisterDataSource("testds1", mockDS1))
	require.NoError(t, reg.RegisterDataSource("testds2", mockDS2))

	funcs := reg.GetFuncs()
	require.Len(t, funcs, 2)
	require.Equal(t, mockFuncDef1, funcs[makeKey("testds1", DataSourceFuncKey("func1"))])
	require.Equal(t, mockFuncDef2, funcs[makeKey("testds2", DataSourceFuncKey("func2"))])
}
