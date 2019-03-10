module Project.Build.Step exposing (Step, decoder)

import Iso8601
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline exposing (custom, required)
import Json.Encode as Encode
import Project.Build.Step.Id as Id exposing (Id)
import Project.Build.Step.Status as Status exposing (Status)
import Time


type Step
    = Step Internals


type alias Internals =
    { id : Id
    , status : Status
    , number :
        Int
        --    , streams : List BuildStream
    , startedAt : Maybe Time.Posix
    , updatedAt : Maybe Time.Posix
    }


decoder : Decoder Step
decoder =
    Decode.succeed Step
        |> custom internalsDecoder


internalsDecoder : Decoder Internals
internalsDecoder =
    Decode.succeed Internals
        |> required "id" Id.decoder
        |> required "status" Status.decoder
        |> required "number" Decode.int
        --        |> required "streams" (Decode.list BuildStream.decoder)
        |>
            required "startedAt" (Decode.maybe Iso8601.decoder)
        |> required "updatedAt" (Decode.maybe Iso8601.decoder)
