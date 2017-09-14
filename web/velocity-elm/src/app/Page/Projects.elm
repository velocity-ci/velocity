module Page.Projects exposing (..)

import Data.Project as Project exposing (Project)
import Data.Session as Session exposing (Session)
import Task exposing (Task)
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Request.Project
import Views.Page as Page
import Http
import Html exposing (..)
import Util exposing ((=>))


-- MODEL --


type alias Model =
    { projects : List Project }


init : Session -> Task PageLoadError Model
init session =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadProjects =
            Request.Project.list maybeAuthToken
                |> Http.toTask

        handleLoadError _ =
            pageLoadError Page.Other "Projects are currently unavailable."
    in
        Task.map Model loadProjects
            |> Task.mapError handleLoadError



-- VIEW --


view : Session -> Model -> Html Msg
view session model =
    div [] [ text "Projects page!" ]



-- UPDATE --


type Msg
    = NoOp


update : Session -> Msg -> Model -> ( Model, Cmd Msg )
update session msg model =
    model => Cmd.none
