module Data.Task exposing (BuildStep, ChoiceParameter, CloneStep, ComposeStep, DerivedParameter, Id(..), Name(..), Parameter(..), PushStep, RunStep, Step(..), StringParameter, Task, basicParameterDecoder, buildStepDecoder, choiceParameterDecoder, cloneStepDecoder, composeStepDecoder, decodeId, decodeName, decoder, derivedParameterDecoder, idEquals, idToString, nameParser, nameToString, parameterDecoder, pushStepDecoder, runStepDecoder, stepDecoder, stepName, stringParameterDecoder)

import Data.Commit as Commit exposing (Commit)
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (decode, optional, required)
import UrlParser


-- MODEL --


type alias Task =
    { id : Id
    , slug : Slug
    , name : Name
    , description : String
    , steps : List Step
    , parameters : List Parameter
    , commit : Commit
    }


type Step
    = Build BuildStep
    | Run RunStep
    | Clone CloneStep
    | Compose ComposeStep
    | Push PushStep


type alias CloneStep =
    { description : String }


type alias PushStep =
    { description : String }


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


type alias ComposeStep =
    { description : String }


type Parameter
    = StringParam StringParameter
    | ChoiceParam ChoiceParameter
    | DerivedParam DerivedParameter


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


type alias DerivedParameter =
    { use : String }



-- SERIALIZATION --


decoder : Decoder Task
decoder =
    decode Task
        |> required "id" decodeId
        |> required "slug" decodeSlug
        |> required "name" decodeName
        |> optional "description" Decode.string ""
        |> optional "steps" (Decode.list stepDecoder) []
        |> optional "parameters" (Decode.list parameterDecoder) []
        |> required "commit" Commit.decoder


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


derivedParameterDecoder : Decoder DerivedParameter
derivedParameterDecoder =
    decode DerivedParameter
        |> required "use" Decode.string


parameterDecoder : Decoder Parameter
parameterDecoder =
    Decode.string
        |> Decode.field "type"
        |> Decode.andThen
            (\paramType ->
                case paramType of
                    "basic" ->
                        basicParameterDecoder

                    "derived" ->
                        Decode.map DerivedParam derivedParameterDecoder

                    unknown ->
                        Decode.fail <| "Unknown parameter type: " ++ unknown
            )


basicParameterDecoder : Decoder Parameter
basicParameterDecoder =
    Decode.string
        |> Decode.list
        |> Decode.nullable
        |> Decode.field "otherOptions"
        |> Decode.andThen
            (\otherOptions ->
                let
                    string =
                        Decode.map StringParam stringParameterDecoder

                    choice =
                        Decode.map ChoiceParam choiceParameterDecoder
                in
                    case otherOptions of
                        Nothing ->
                            string

                        Just options ->
                            if List.isEmpty options then
                                string
                            else
                                choice
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

                    "setup" ->
                        Decode.map Clone cloneStepDecoder

                    "compose" ->
                        Decode.map Compose composeStepDecoder

                    "push" ->
                        Decode.map Push pushStepDecoder

                    unknown ->
                        Decode.fail <| "Unknown type: " ++ unknown
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


composeStepDecoder : Decoder ComposeStep
composeStepDecoder =
    decode ComposeStep
        |> required "description" Decode.string


pushStepDecoder : Decoder ComposeStep
pushStepDecoder =
    decode PushStep
        |> required "description" Decode.string



-- IDENTIFIERS --


type Name
    = Name String


type Id
    = Id String

type Slug
    = Slug String


decodeName : Decoder Name
decodeName =
    Decode.map Name Decode.string

decodeSlug : Decoder Slug
decodeSlug =
    Decode.map Slug Decode.string


nameParser : UrlParser.Parser (Name -> a) a
nameParser =
    UrlParser.custom "NAME" (Ok << Name)


nameToString : Name -> String
nameToString (Name name) =
    name

slugParser : UrlParser.Parser (Slug -> a) a
slugParser =
    UrlParser.custom "SLUG" (Ok << Slug)

slugToString : Slug -> String
slugToString (Slug slug) =
    slug


decodeId : Decoder Id
decodeId =
    Decode.map Id Decode.string


idToString : Id -> String
idToString (Id id) =
    id


idEquals : Id -> Id -> Bool
idEquals (Id first) (Id second) =
    first == second



-- HELPERS --


stepName : Step -> String
stepName taskStep =
    case taskStep of
        Build _ ->
            "Build"

        Run _ ->
            "Run"

        Clone _ ->
            "Clone"

        Compose _ ->
            "Compose"

        Push _ ->
            "Push"
