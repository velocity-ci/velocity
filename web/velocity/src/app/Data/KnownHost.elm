module Data.KnownHost exposing (KnownHost, decoder, addKnownHost)

import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (custom, decode, hardcoded, required)


type alias KnownHost =
    { id : Id
    , hosts : List String
    , comment : String
    , sha256 : String
    , md5 : String
    }



-- SERIALIZATION --


decoder : Decoder KnownHost
decoder =
    decode KnownHost
        |> required "id" decodeId
        |> required "hosts" (Decode.list Decode.string)
        |> required "comment" Decode.string
        |> required "sha256" Decode.string
        |> required "md5" Decode.string



-- HELPERS --


findKnownHost : List KnownHost -> KnownHost -> Maybe KnownHost
findKnownHost knownHosts knownHost =
    List.filter (\a -> a.id == knownHost.id) knownHosts
        |> List.head


addKnownHost : List KnownHost -> KnownHost -> List KnownHost
addKnownHost knownHosts knownHost =
    case findKnownHost knownHosts knownHost of
        Just _ ->
            knownHosts

        Nothing ->
            knownHost :: knownHosts



-- IDENTIFIERS --


type Id
    = Id String


decodeId : Decoder Id
decodeId =
    Decode.map Id Decode.string


idToString : Id -> String
idToString (Id id) =
    id
