module Data.Project exposing (Id(..), Project, Slug(..), addProject, decodeId, decoder, idParser, idToString, slugParser, slugToString)

import Data.Helpers exposing (stringToDateTime)
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (custom, hardcoded, optional, required)
import Time.DateTime as DateTime exposing (DateTime)
import UrlParser
import Iso8601


type alias Project =
    { id : Id
    , slug : Slug
    , name : String
    , repository : String
    , createdAt : DateTime
    , updatedAt : DateTime
    , synchronising : Bool
    , logo : Maybe String
    }



-- SERIALIZATION --


decoder : Decoder Project
decoder =
    Decode.succeed Project
        |> required "id" decodeId
        |> required "slug" decodeSlug
        |> required "name" Decode.string
        |> required "repository" Decode.string
        |> required "createdAt" Iso8601.decoder
        |> required "updatedAt" Iso8601.decoder
        |> required "synchronising" Decode.bool
        |> required "logo" (Decode.maybe Decode.string)



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
