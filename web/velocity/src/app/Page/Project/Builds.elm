module Page.Project.Builds exposing (..)

-- EXTERNAL --

import Html exposing (..)
import Http
import Task exposing (Task)


-- INTERNAL --

import Context exposing (Context)
import Data.Build as Build exposing (Build)
import Data.Project as Project exposing (Project)
import Data.PaginatedList as PaginatedList exposing (PaginatedList)
import Data.Session as Session exposing (Session)
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Request.Errors
import Request.Project
import Util exposing ((=>))
import Views.Page as Page


-- MODEL --


type alias Model =
    { builds : PaginatedList Build }


init : Context -> Session msg -> Project.Slug -> Maybe Int -> Task PageLoadError Model
init context session slug maybePage =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadBuilds =
            Request.Project.builds context slug maybeAuthToken

        handleLoadError _ =
            pageLoadError Page.Project "Project unavailable."
    in
        Task.map Model loadBuilds
            |> Task.mapError handleLoadError



-- VIEW --


view : Model -> Html Msg
view model =
    text "Builds page"



-- UPDATE --


type Msg
    = NoOp


update : Context -> Project -> Session msg -> Msg -> Model -> ( Model, Cmd Msg )
update context project session msg model =
    model => Cmd.none
