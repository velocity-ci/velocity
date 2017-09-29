module Data.Task exposing (..)

import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (decode, required, optional)
import UrlParser


-- MODEL --


type alias Task =
    { name : Name
    , description : String
    , steps : List Step
    }


type Step
    = Build BuildStep
    | Run RunStep


type alias BuildStep =
    { description : String
    , dockerfile : String
    , context : String
    , tags : List String
    }


type alias RunStep =
    { description : String
    , command : List String
    , environment : List ( String, String )
    , ignoreExitCode : Bool
    , image : String
    , mountPoint : String
    , workingDir : String
    }



-- SERIALIZATION --


decoder : Decoder Task
decoder =
    decode Task
        |> required "name" (Decode.map Name Decode.string)
        |> optional "description" Decode.string ""
        |> optional "steps" (Decode.list stepDecoder) []


stepDecoder : Decoder Step
stepDecoder =
    Decode.field "type" Decode.string
        |> Decode.andThen
            (\type_ ->
                case type_ of
                    "build" ->
                        Decode.map Build buildStepDecoder

                    "run" ->
                        Decode.map Run runStepDecoder

                    unknownType ->
                        Decode.fail <| "Unknown type: " ++ unknownType
            )


buildStepDecoder : Decoder BuildStep
buildStepDecoder =
    decode BuildStep
        |> required "description" Decode.string
        |> required "dockerfile" Decode.string
        |> required "context" Decode.string
        |> required "tags" (Decode.list Decode.string)


runStepDecoder : Decoder RunStep
runStepDecoder =
    decode RunStep
        |> required "description" Decode.string
        |> optional "command" (Decode.list Decode.string) []
        |> optional "environment" (Decode.keyValuePairs Decode.string) []
        |> optional "ignoreExitCode" Decode.bool False
        |> required "image" Decode.string
        |> required "mountPoint" Decode.string
        |> required "workingDir" Decode.string



-- IDENTIFIERS --


type Name
    = Name String


nameParser : UrlParser.Parser (Name -> a) a
nameParser =
    UrlParser.custom "NAME" (Ok << Name)


nameToString : Name -> String
nameToString (Name name) =
    name
