module Test.Context exposing (..)

import Expect exposing (Expectation)
import Test exposing (..)
import Context


suite : Test
suite =
    describe "Context"
        [ describe "initContext"
            [ describe "Creates correct websocket URL"
                [ test "http:" <|
                    \() ->
                        Expect.equal
                            (Context.initContext "http://example.com/api/v1")
                            { apiUrlBase = "http://example.com/api/v1"
                            , wsUrl = "ws://example.com/api/v1/ws"
                            }
                , test "https:" <|
                    \() ->
                        Expect.equal
                            (Context.initContext "https://example.com/api/v1")
                            { apiUrlBase = "https://example.com/api/v1"
                            , wsUrl = "wss://example.com/api/v1/ws"
                            }
                ]
            ]
        ]
