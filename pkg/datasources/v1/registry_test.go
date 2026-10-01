// SPDX-FileCopyrightText: Copyright 2024 The Minder Authors
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/mindersec/minder/pkg/engine/v1/interfaces"
)

type dummyDataSourceFuncDef struct{}

func (d *dummyDataSourceFuncDef) ValidateArgs(obj any) error { return nil }
func (d *dummyDataSourceFuncDef) ValidateUpdate(obj *structpb.Struct) error { return nil }
func (d *dummyDataSourceFuncDef) Call(ctx context.Context, ingest *interfaces.Ingested, args any) (any, error) { return nil, nil }
func (d *dummyDataSourceFuncDef) GetArgsSchema() *structpb.Struct { return nil }

type dummyDataSource struct {
	funcs map[DataSourceFuncKey]DataSourceFuncDef
}

func (d *dummyDataSource) GetFuncs() map[DataSourceFuncKey]DataSourceFuncDef {
	return d.funcs
}

func TestDataSourceRegistry_RegisterDataSource(t *testing.T) {
	t.Parallel()

	dummyDef := &dummyDataSourceFuncDef{}
	
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
				DataSourceFuncKey("func1"): dummyDef,
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reg := NewDataSourceRegistry()
			ds := &dummyDataSource{funcs: tc.funcs}

			err := reg.RegisterDataSource(tc.dsName, ds)
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

	dummyDef := &dummyDataSourceFuncDef{}
	reg := NewDataSourceRegistry()
	ds := &dummyDataSource{
		funcs: map[DataSourceFuncKey]DataSourceFuncDef{
			DataSourceFuncKey("func1"): dummyDef,
		},
	}

	err := reg.RegisterDataSource("testds", ds)
	require.NoError(t, err)

	// Registering again with same name and same function key should fail
	err = reg.RegisterDataSource("testds", ds)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrDuplicateDataSourceFuncKey)
}

func TestDataSourceRegistry_GetFuncs(t *testing.T) {
	t.Parallel()

	dummyDef1 := &dummyDataSourceFuncDef{}
	dummyDef2 := &dummyDataSourceFuncDef{}

	reg := NewDataSourceRegistry()
	ds1 := &dummyDataSource{
		funcs: map[DataSourceFuncKey]DataSourceFuncDef{
			DataSourceFuncKey("func1"): dummyDef1,
		},
	}
	ds2 := &dummyDataSource{
		funcs: map[DataSourceFuncKey]DataSourceFuncDef{
			DataSourceFuncKey("func2"): dummyDef2,
		},
	}

	require.NoError(t, reg.RegisterDataSource("testds1", ds1))
	require.NoError(t, reg.RegisterDataSource("testds2", ds2))

	funcs := reg.GetFuncs()
	require.Len(t, funcs, 2)
	require.Equal(t, dummyDef1, funcs[makeKey("testds1", DataSourceFuncKey("func1"))])
	require.Equal(t, dummyDef2, funcs[makeKey("testds2", DataSourceFuncKey("func2"))])
}
