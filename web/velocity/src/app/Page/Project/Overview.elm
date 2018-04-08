module Page.Project.Overview exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Data.Project as Project exposing (Project)
import Page.Helpers exposing (formatDateTime)
import Data.Build as Build exposing (Build)
import Views.Build exposing (viewBuildStatusIcon, viewBuildTextClass)
import Util exposing ((=>))
import Data.Session as Session exposing (Session)
import Navigation
import Route
import Page.Project.Route as ProjectRoute
import Page.Project.Commit.Route as CommitRoute
import Views.Helpers exposing (onClickPage)
import Data.Task as Task
import Data.Commit as Commit


-- MODEL --


type alias Model =
    {}


initialModel : Model
initialModel =
    {}



-- VIEW --


view : Project -> List Build -> Html Msg
view project builds =
    div [ class "row" ]
        [ viewOverviewCard project
        , viewBuildHistoryTable project builds
        ]


viewOverviewCard : Project -> Html Msg
viewOverviewCard project =
    div [ class "col-md-6" ]
        [ div [ class "card mb-3" ]
            [ div [ class "card-body" ]
                [ dl [ class "mb-0" ]
                    [ dt [] [ text "Repository" ]
                    , dd [] [ text project.repository ]
                    , dt [] [ text "Last update" ]
                    , dd [] [ text (formatDateTime project.updatedAt) ]
                    ]
                ]
            ]
        ]


viewBuildHistoryTable : Project -> List Build -> Html Msg
viewBuildHistoryTable project builds =
    div [ class "col-md-6" ]
        [ div [ class "card" ]
            [ h5 [ class "card-header border-bottom-0" ] [ text "Build history" ]
            , table [ class "table mb-0 " ] (List.map (viewBuildHistoryTableRow project) (List.take 10 builds))
            ]
        ]


viewBuildHistoryTableRow : Project -> Build -> Html Msg
viewBuildHistoryTableRow project build =
    let
        colourClassList =
            [ viewBuildTextClass build => True ]

        commitTaskRoute =
            CommitRoute.Task build.task.name Nothing
                |> ProjectRoute.Commit build.task.commit.hash
                |> Route.Project project.slug

        commitRoute =
            CommitRoute.Overview
                |> ProjectRoute.Commit build.task.commit.hash
                |> Route.Project project.slug

        task =
            build.task

        taskName =
            Task.nameToString task.name

        createdAt =
            formatDateTime build.createdAt

        truncatedHash =
            Commit.truncateHash task.commit.hash

        buildLink content route =
            a
                [ Route.href route
                , onClickPage NewUrl route
                , classList colourClassList
                ]
                [ text content ]
    in
        tr [ classList colourClassList ]
            [ td [] [ buildLink taskName commitTaskRoute ]
            , td [] [ buildLink truncatedHash commitRoute ]
            , td [] [ buildLink createdAt commitTaskRoute ]
            , td [ class "text-right" ] [ viewBuildStatusIcon build ]
            ]



-- UPDATE --


type Msg
    = NewUrl String


update : Project -> Session msg -> Msg -> Model -> ( Model, Cmd Msg )
update project session msg model =
    case msg of
        NewUrl url ->
            model => Navigation.newUrl url
