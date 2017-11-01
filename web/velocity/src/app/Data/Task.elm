module Data.Task exposing (..)

import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (decode, optional, required)
import UrlParser


-- MODEL --


type alias Task =
    { name : Name
    , description : String
    , steps : List Step
    , parameters : List Parameter
    }


type Step
    = Build BuildStep
    | Run RunStep
    | Clone CloneStep


type alias CloneStep =
    { description : String
    , submodule : Bool
    }


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


type Parameter
    = StringParam StringParameter
    | ChoiceParam ChoiceParameter


type alias StringParameter =
    { name : String
    , default : Maybe String
    , secret : Bool
    }


type alias ChoiceParameter =
    { name : String
    , default : Maybe String
    , secret : Bool
    , options : List String
    }



-- SERIALIZATION --


decoder : Decoder Task
decoder =
    decode Task
        |> required "name" (Decode.map Name Decode.string)
        |> optional "description" Decode.string ""
        |> optional "steps" (Decode.list stepDecoder) []
        |> optional "parameters" parameterKeyValuePairDecoder []


parameterKeyValuePairDecoder : Decoder (List Parameter)
parameterKeyValuePairDecoder =
    Decode.keyValuePairs parameterOptionsDecoder
        |> Decode.andThen (List.map Tuple.second >> Decode.succeed)


stringParameterDecoder : Decoder StringParameter
stringParameterDecoder =
    decode StringParameter
        |> required "name" Decode.string
        |> optional "default" (Decode.nullable Decode.string) Nothing
        |> optional "secret" Decode.bool False


choiceParameterDecoder : Decoder ChoiceParameter
choiceParameterDecoder =
    decode ChoiceParameter
        |> required "name" Decode.string
        |> optional "default" (Decode.nullable Decode.string) Nothing
        |> optional "secret" Decode.bool False
        |> required "otherOptions" (Decode.list Decode.string)


parameterOptionsDecoder : Decoder Parameter
parameterOptionsDecoder =
    Decode.string
        |> Decode.list
        |> Decode.nullable
        |> Decode.field "otherOptions"
        |> Decode.andThen
            (\otherOptions ->
                case otherOptions of
                    Nothing ->
                        Decode.map StringParam stringParameterDecoder

                    Just options ->
                        Decode.map ChoiceParam choiceParameterDecoder
            )


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

                    "clone" ->
                        Decode.map Clone cloneStepDecoder

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


cloneStepDecoder : Decoder CloneStep
cloneStepDecoder =
    decode CloneStep
        |> required "description" Decode.string
        |> required "submodule" Decode.bool



-- IDENTIFIERS --


type Name
    = Name String


nameParser : UrlParser.Parser (Name -> a) a
nameParser =
    UrlParser.custom "NAME" (Ok << Name)


nameToString : Name -> String
nameToString (Name name) =
    name
