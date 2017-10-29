module Decoders.Commit exposing (..)

import Expect exposing (Expectation)
import Fuzz exposing (Fuzzer, int, list, string)
import Test exposing (..)
import Json.Decode as Decode
import Data.Branch as Branch
import Data.Commit as Commit


suite : Test
suite =
    describe "decoders"
        [ describe "commits"
            [ test "properly decodes" <|
                \() ->
                    let
                        data =
                            """
                            "develop"
                            """
                    in
                        Expect.equal (Decode.decodeString Branch.decoder data) (Ok <| Branch.Name "develop")
            ]
        ]
