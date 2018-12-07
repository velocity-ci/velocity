module Project.Task.Step exposing (Step, decoder)

import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline exposing (custom, optional, required)
import Json.Encode as Encode


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


decoder : Decoder Step
decoder =
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
    Decode.succeed BuildStep
        |> required "description" Decode.string
        |> required "dockerfile" Decode.string
        |> required "context" Decode.string
        |> required "tags" (Decode.list Decode.string)


runStepDecoder : Decoder RunStep
runStepDecoder =
    Decode.succeed RunStep
        |> required "description" Decode.string
        |> optional "command" (Decode.list Decode.string) []
        |> optional "environment" (Decode.keyValuePairs Decode.string) []
        |> optional "ignoreExitCode" Decode.bool False
        |> required "image" Decode.string
        |> required "mountPoint" Decode.string
        |> required "workingDir" Decode.string


cloneStepDecoder : Decoder CloneStep
cloneStepDecoder =
    Decode.succeed CloneStep
        |> required "description" Decode.string


composeStepDecoder : Decoder ComposeStep
composeStepDecoder =
    Decode.succeed ComposeStep
        |> required "description" Decode.string


pushStepDecoder : Decoder ComposeStep
pushStepDecoder =
    Decode.succeed PushStep
        |> required "description" Decode.string
