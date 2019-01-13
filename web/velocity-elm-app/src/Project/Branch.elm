module Project.Branch exposing (Branch, default, text)

import Element exposing (..)
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline exposing (custom, required)
import Project.Branch.Name as Name exposing (Name)


type Branch
    = Branch Internals


type alias Internals =
    { name : Name
    , active : Bool
    }


default : Branch
default =
    Branch
        { name = Name.default
        , active = True
        }



-- INFO


text : Branch -> Element msg
text (Branch { name }) =
    Name.text name



-- SERIALIZATION --


decoder : Decoder Branch
decoder =
    Decode.succeed Branch
        |> custom internalsDecoder


internalsDecoder : Decoder Internals
internalsDecoder =
    Decode.succeed Internals
        |> required "name" Name.decoder
        |> required "active" Decode.bool
