module Page.Projects exposing (..)

import Data.Project as Project exposing (Project)
import Data.Session as Session exposing (Session)
import Task exposing (Task)
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Request.Project
import Views.Page as Page
import Http
import Html exposing (..)
import Html.Attributes exposing (..)
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
    div [ class "row", attribute "style" "margin-top:3em;" ]
        [ div [ class "col-12" ]
            [ div [ class "card" ]
                [ h4 [ class "card-header" ]
                    [ text ("Projects (" ++ (model.projects |> List.length |> toString) ++ ")") ]
                , ul [ class "list-group" ] (List.map viewProject model.projects)
                ]
            ]
        ]


viewProject : Project -> Html Msg
viewProject project =
    li [ class "list-group-item list-group-item-action flex-column align-items-start" ]
        [ div [ class "d-flex w-100 justify-content-between" ]
            [ h5 [ class "mb-1" ]
                [ a [ href "#" ]
                    [ text project.name ]
                ]
            , small []
                [ text project.updatedAt ]
            ]
        , p [ class "mb-1" ]
            [ text "Donec id elit non mi porta gravida at eget metus. Maecenas sed diam eget risus varius blandit." ]
        , small []
            [ text "Donec id elit non mi porta." ]
        ]



-- UPDATE --


type Msg
    = NoOp


update : Session -> Msg -> Model -> ( Model, Cmd Msg )
update session msg model =
    model => Cmd.none
