/**
 * Copyright (c) 2022, Xerra Earth Observation Institute
 * All rights reserved. Use is subject to License terms.
 * See LICENSE.TXT in the root directory of this source tree.
 */

package lockunique_test

import (
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/rs/zerolog/log"
)

func RunDebugServer(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/{action}", pprof.Index)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	log.Info().Msgf("Starting debug server at http://localhost:%d/debug/pprof/", port)
	err := http.ListenAndServe(fmt.Sprintf("localhost:%d", port), mux)
	log.Error().Err(err).Msg("debug http server exited")
}
