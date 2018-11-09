module Page.Home.ActivePanel exposing (ActivePanel(..), queryParser, toQueryParams)

import Json.Decode as Decode exposing (Decoder)
import Url.Builder
import Url.Parser.Query as QueryParser


type ActivePanel
    = NewProjectForm


toQueryParams : String -> Maybe ActivePanel -> List Url.Builder.QueryParameter
toQueryParams key maybeActivePanel =
    case maybeActivePanel of
        Just NewProjectForm ->
            [ Url.Builder.string key "new-project" ]

        Nothing ->
            []


queryParser : String -> QueryParser.Parser (Maybe ActivePanel)
queryParser key =
    QueryParser.custom key <|
        \stringList ->
            case stringList of
                [ "new-project" ] ->
                    Just NewProjectForm

                _ ->
                    Nothing
