module Decoders.Project exposing (..)

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
        [ describe "project"
            [ test "properly decodes when no key is present " <|
                \() ->
                    let
                        inputMissingKey =
                            """
                            {
                              "name":"Key test",
                              "repository":"https://github.com/velocity-ci/velocity",
                              "id":"key-test",
                              "createdAt":"2017-10-01T14:23:06+0100",
                              "updatedAt":"2017-10-03T09:23:06+0100",
                              "synchronising":false
                            }
                            """

                        inputNullKey =
                            """
                            {
                              "name":"Key test",
                              "key": null,
                              "repository":"https://github.com/velocity-ci/velocity",
                              "id":"key-test",
                              "createdAt":"2017-10-01T14:23:06+0100",
                              "updatedAt":"2017-10-03T09:23:06+0100",
                              "synchronising":false
                            }
                            """
                    in
                        case ( fromISO8601 "2017-10-01T14:23:06+0100", fromISO8601 "2017-10-03T09:23:06+0100" ) of
                            ( Ok createdAt, Ok updatedAt ) ->
                                let
                                    expected =
                                        (Ok
                                            { id = Project.Id "key-test"
                                            , key = Nothing
                                            , name = "Key test"
                                            , repository = "https://github.com/velocity-ci/velocity"
                                            , createdAt = createdAt
                                            , updatedAt = updatedAt
                                            }
                                        )
                                in
                                    ( decodeStringToProject inputMissingKey, decodeStringToProject inputNullKey )
                                        |> Expect.equal ( expected, expected )

                            ( _, _ ) ->
                                Expect.fail "Could not parse datetimes"
            , test "properly decodes when key is present" <|
                \() ->
                    let
                        input =
                            """
                            {
                              "name":"Key test",
                              "repository":"https://github.com/velocity-ci/velocity",
                              "id":"key-test",
                              "key": "--- BEGIN RSA PRIVATE KEY --- ...",
                              "createdAt":"2017-10-01T14:23:06+0100",
                              "updatedAt":"2017-10-03T09:23:06+0100",
                              "synchronising":false
                            }
                            """
                    in
                        case ( fromISO8601 "2017-10-01T14:23:06+0100", fromISO8601 "2017-10-03T09:23:06+0100" ) of
                            ( Ok createdAt, Ok updatedAt ) ->
                                decodeStringToProject input
                                    |> Expect.equal
                                        (Ok
                                            { id = Project.Id "key-test"
                                            , key = Just "--- BEGIN RSA PRIVATE KEY --- ..."
                                            , name = "Key test"
                                            , repository = "https://github.com/velocity-ci/velocity"
                                            , createdAt = createdAt
                                            , updatedAt = updatedAt
                                            }
                                        )

                            ( _, _ ) ->
                                Expect.fail "Could not parse datetimes"
            ]
        ]


decodeStringToProject : String -> Result String Project
decodeStringToProject input =
    input
        |> Decode.decodeString Project.decoder
