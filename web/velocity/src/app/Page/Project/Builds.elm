module Page.Project.Builds exposing (Model, Msg(..), init, pageLink, pagination, perPage, update, view)

-- EXTERNAL --
-- INTERNAL --

import Context exposing (Context)
import Data.Build as Build exposing (Build)
import Data.PaginatedList as PaginatedList exposing (PaginatedList)
import Data.Project as Project exposing (Project)
import Data.Session as Session exposing (Session)
import Html exposing (..)
import Html.Attributes exposing (..)
import Navigation
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Page.Project.Route as ProjectRoute
import Request.Project
import Route
import Task exposing (Task)
import Util exposing ((=>))
import Views.Build exposing (viewBuildHistoryTable)
import Views.Helpers exposing (onClickPage)
import Views.Page as Page


-- MODEL --


type alias Model =
    { builds : PaginatedList Build
    , activePage : Int
    }


init : Context -> Session msg -> Project.Slug -> Maybe Int -> Task PageLoadError Model
init context session slug maybePage =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadBuilds =
            Request.Project.builds context slug maybeAuthToken perPage (Just activePage)

        handleLoadError _ =
            pageLoadError Page.Project "Project unavailable."

        activePage =
            Maybe.withDefault 1 maybePage
    in
        Task.map2 Model loadBuilds (Task.succeed activePage)
            |> Task.mapError handleLoadError


perPage : Int
perPage =
    15



-- VIEW --


view : Project -> Model -> Html Msg
view project model =
    let
        builds =
            PaginatedList.results model.builds

        total =
            PaginatedList.total model.builds
    in
        div []
            [ h4 [ class "mb-4" ] [ text "Builds" ]
            , div [ class "mb-4" ] [ viewBuildHistoryTable project builds NewUrl ]
            , pagination model.activePage total project
            ]


pagination : Int -> Int -> Project -> Html Msg
pagination activePage total project =
    let
        totalPages =
            ceiling (toFloat total / toFloat perPage)
    in
        if totalPages > 1 then
            List.range 1 totalPages
                |> List.map (\page -> pageLink page (page == activePage) project)
                |> ul [ class "pagination" ]
        else
            Html.text ""


pageLink : Int -> Bool -> Project -> Html Msg
pageLink page isActive project =
    let
        route =
            Route.Project project.slug <| ProjectRoute.Builds (Just page)
    in
        li [ classList [ "page-item" => True, "active" => isActive ] ]
            [ a
                [ class "page-link"
                , Route.href route
                , onClickPage NewUrl route
                ]
                [ text (toString page) ]
            ]



-- UPDATE --


type Msg
    = NewUrl String
    | UpdateBuild Build
    | AddBuild Build


update : Context -> Project -> Session msg -> Msg -> Model -> ( Model, Cmd Msg )
update context project session msg model =
    case msg of
        NewUrl url ->
            model => Navigation.newUrl url

        AddBuild build ->
            let
                newModel =
                    if model.activePage == 1 then
                        { model | builds = PaginatedList.updateResults model.builds (\builds -> build :: builds) }
                    else
                        model
            in
                newModel => Cmd.none

        UpdateBuild build ->
            let
                builds =
                    PaginatedList.updateResults model.builds
                        (List.map
                            (\a ->
                                if build.id == a.id then
                                    build
                                else
                                    a
                            )
                        )
            in
                { model | builds = builds }
                    => Cmd.none
