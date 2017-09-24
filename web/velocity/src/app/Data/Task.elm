module Data.Task exposing (..)

import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (decode, required, optional)


-- MODEL --


type alias Task =
    { name : String
    , description : String
    }


type StepType
    = Build
    | Run


type alias Step =
    { type_ : String
    , dockerfile : Maybe String
    , tags : Maybe (List String)
    , description : Maybe String
    , context : Maybe String
    , workingDir : Maybe String
    , mountPoint : Maybe String
    , ignoreExitCode : Maybe Bool
    , command : Maybe (List String)
    }



-- SERIALIZATION --


decoder : Decoder Task
decoder =
    decode Task
        |> required "name" Decode.string
        |> optional "description" Decode.string ""
