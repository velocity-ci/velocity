module Page.Project.Overview exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Data.Project as Project exposing (Project)
import Page.Helpers exposing (formatDateTime)
import Data.Build as Build exposing (Build)
import Views.Build exposing (viewBuildStatusIcon, viewBuildTextClass)
import Util exposing ((=>))
import Data.Session as Session exposing (Session)
import Views.Helpers exposing (onClickPage)
import Navigation
import Route exposing (Route)
import Page.Project.Route as ProjectRoute
import Page.Project.Commit.Route as CommitRoute


-- MODEL --


type alias Model =
    {}


initialModel : Model
initialModel =
    {}



-- VIEW --


view : Project -> List Build -> Html msg
view project builds =
    div []
        [ viewOverviewCard project
        , viewBuildHistoryTable project builds
        ]


viewOverviewCard : Project -> Html msg
viewOverviewCard project =
    div [ class "card mb-3" ]
        [ div [ class "card-body" ]
            [ dl [ class "mb-0" ]
                [ dt [] [ text "Repository" ]
                , dd [] [ text project.repository ]
                , dt [] [ text "Last update" ]
                , dd [] [ text (formatDateTime project.updatedAt) ]
                ]
            ]
        ]


viewBuildHistoryTable : Project -> List Build -> Html msg
viewBuildHistoryTable project builds =
    div [ class "card" ]
        [ h5 [ class "card-header border-bottom-0" ] [ text "Build history" ]
        , table [ class "table mb-0 " ] (List.map (viewBuildHistoryTableRow project) (List.take 10 builds))
        ]


viewBuildHistoryTableRow : Project -> Build -> Html msg
viewBuildHistoryTableRow project build =
    let
        rowClasses =
            [ (viewBuildTextClass build) => True ]

        --
        --        route =
        --            CommitRoute.Task build.taskId Nothing
        --                |> ProjectRoute.Commit build.commitHash
        --                |> Route.Project project.id
    in
        tr [ classList rowClasses ]
            [ td [ class "d-flex justify-content-between" ]
                [ text ((formatDateTime build.createdAt) ++ " ")
                , viewBuildStatusIcon build
                ]
            ]



-- UPDATE --


type Msg
    = NewUrl String


update : Project -> Session msg -> Msg -> Model -> ( Model, Cmd Msg )
update project session msg model =
    case msg of
        NewUrl url ->
            model => Navigation.newUrl url
