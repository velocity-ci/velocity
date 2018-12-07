module Project.Branch exposing (Branch)

import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline exposing (custom, required)
import Json.Encode as Encode
import Project.Branch.Name as Name exposing (Name)


type Branch
    = Branch Internals


type alias Internals =
    { name : Name
    , active : Bool
    }



-- SERIALIZATION --


decoder : Decoder Branch
decoder =
    Decode.succeed Branch
        |> required "name" Name.decoder
        |> required "active" Decode.bool
