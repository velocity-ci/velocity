module Data.Project exposing (Project, decoder, idParser, slugParser, idToString, slugToString, decodeId, Id(..), Slug(..), addProject)

import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (custom, decode, hardcoded, required, optional)
import Time.DateTime as DateTime exposing (DateTime)
import Data.Helpers exposing (stringToDateTime)
import UrlParser


type alias Project =
    { id : Id
    , slug : Slug
    , name : String
    , repository : String
    , createdAt : DateTime
    , updatedAt : DateTime
    , synchronising : Bool
    }



-- SERIALIZATION --


decoder : Decoder Project
decoder =
    decode Project
        |> required "id" decodeId
        |> required "slug" decodeSlug
        |> required "name" Decode.string
        |> required "repository" Decode.string
        |> required "createdAt" stringToDateTime
        |> required "updatedAt" stringToDateTime
        |> required "synchronising" Decode.bool



-- HELPERS --


findProject : List Project -> Project -> Maybe Project
findProject projects project =
    List.filter (\a -> a.id == project.id) projects
        |> List.head


addProject : List Project -> Project -> List Project
addProject projects project =
    case findProject projects project of
        Just _ ->
            projects

        Nothing ->
            project :: projects



-- IDENTIFIERS --


type Id
    = Id String


type Slug
    = Slug String


decodeId : Decoder Id
decodeId =
    Decode.map Id Decode.string


decodeSlug : Decoder Slug
decodeSlug =
    Decode.map Slug Decode.string


idParser : UrlParser.Parser (Id -> a) a
idParser =
    UrlParser.custom "ID" (Ok << Id)


slugParser : UrlParser.Parser (Slug -> a) a
slugParser =
    UrlParser.custom "SLUG" (Ok << Slug)


idToString : Id -> String
idToString (Id id) =
    id


slugToString : Slug -> String
slugToString (Slug slug) =
    slug
