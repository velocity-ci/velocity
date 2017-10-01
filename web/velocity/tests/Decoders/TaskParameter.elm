module Decoders.TaskParameter exposing (..)

import Expect exposing (Expectation)
import Fuzz exposing (Fuzzer, int, list, string)
import Test exposing (..)
import Json.Decode as Decode
import Data.Project as Project exposing (Project)
import Data.Task as ProjectTask
import Time.DateTime as DateTime exposing (fromISO8601)


suite : Test
suite =
    describe "decoders"
        [ describe "task parameters"
            [ test "properly decodes a simple task parameter with a default" <|
                \() ->
                    let
                        input =
                            """
                                      {
                                        "name":"environment",
                                        "default":"testing",
                                        "secret":false
                                      }
                                      """
                    in
                        input
                            |> Decode.decodeString ProjectTask.parameterDecoder
                            |> Expect.equal
                                (Ok
                                    { name = "environment"
                                    , default = Just "testing"
                                    , secret = False
                                    }
                                )
            , test "properly decodes a simple task parameter with no default" <|
                \() ->
                    let
                        input =
                            """
                                      {
                                        "name":"environment",
                                        "secret":true
                                      }
                                      """
                    in
                        input
                            |> Decode.decodeString ProjectTask.parameterDecoder
                            |> Expect.equal
                                (Ok
                                    { name = "environment"
                                    , default = Nothing
                                    , secret = True
                                    }
                                )
            ]
        ]
