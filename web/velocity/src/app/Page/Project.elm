module Page.Project exposing (..)

import Context exposing (Context)
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Request.Project
import Request.Errors
import Task exposing (Task)
import Data.Session as Session exposing (Session)
import Data.Project as Project exposing (Project)
import Data.Branch as Branch exposing (Branch)
import Html exposing (..)
import Html.Attributes exposing (..)
import Views.Page as Page
import Util exposing ((=>), viewIf)
import Http
import Route exposing (Route)
import Views.Page as Page exposing (ActivePage)
import Page.Project.Route as ProjectRoute
import Page.Project.Commits as Commits
import Page.Project.Settings as Settings
import Page.Project.Commit as Commit
import Page.Project.Overview as Overview
import Views.Helpers exposing (onClickPage)
import Navigation exposing (newUrl)
import Data.PaginatedList as PaginatedList exposing (Paginated(..), PaginatedList)
import Json.Encode as Encode
import Json.Decode as Decode
import Dict exposing (Dict)
import Data.Build as Build exposing (Build, addBuild)
import Page.Helpers exposing (sortByDatetime)
import Bootstrap.Popover as Popover
import Component.Sidebar as Sidebar exposing (ActiveSubPage(..))


-- SUB PAGES --


type SubPage
    = Blank
    | Overview Overview.Model
    | Commits Commits.Model
    | Commit Commit.Model
    | Settings Settings.Model
    | Errored PageLoadError


type SubPageState
    = Loaded SubPage
    | TransitioningFrom SubPage



-- MODEL --


type alias Model =
    { project : Project
    , branches : List Branch
    , subPageState : SubPageState
    , builds : List Build
    , commitIconPopover : Popover.State
    , settingsIconPopover : Popover.State
    }


initialSubPage : SubPage
initialSubPage =
    Blank


init : Context -> Session msg -> Project.Slug -> Maybe ProjectRoute.Route -> Task (Request.Errors.Error PageLoadError) ( Model, Cmd Msg )
init context session slug maybeRoute =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadProject =
            Request.Project.get context slug maybeAuthToken

        loadBranches =
            Request.Project.branches context slug maybeAuthToken

        loadBuilds =
            Request.Project.builds context slug maybeAuthToken

        initialModel project (Paginated branches) (Paginated builds) =
            { project = project
            , branches = branches.results
            , subPageState = Loaded initialSubPage
            , builds = builds.results
            , commitIconPopover = Popover.initialState
            , settingsIconPopover = Popover.initialState
            }

        handleLoadError e =
            case e of
                Http.BadPayload debugError _ ->
                    pageLoadError Page.Project debugError

                _ ->
                    pageLoadError Page.Project "Project unavailable."
    in
        Task.map3 initialModel loadProject loadBranches loadBuilds
            |> Task.map
                (\successModel ->
                    case maybeRoute of
                        Just route ->
                            update context session (SetRoute maybeRoute) successModel

                        Nothing ->
                            ( successModel, Cmd.none )
                )
            |> Task.mapError (Request.Errors.mapUnhandledError handleLoadError)



-- CHANNELS --


channelName : Project.Slug -> String
channelName projectSlug =
    "project:" ++ (Project.slugToString projectSlug)


initialEvents : Project.Slug -> ProjectRoute.Route -> Dict String (List ( String, Encode.Value -> Msg ))
initialEvents slug route =
    let
        mapEvents fromMsg events =
            events
                |> Dict.map (\_ v -> List.map (Tuple.mapSecond (\msg -> msg >> fromMsg)) v)

        subPageEvents =
            case route of
                ProjectRoute.Commits _ _ ->
                    mapEvents CommitsMsg (Commits.events slug)

                ProjectRoute.Commit _ subRoute ->
                    mapEvents CommitMsg (Commit.initialEvents slug subRoute)

                _ ->
                    Dict.empty

        pageEvents =
            [ ( "project:update", UpdateProject )
            , ( "project:delete", ProjectDeleted )
            , ( "branch:new", RefreshBranches )
            , ( "branch:update", RefreshBranches )
            , ( "branch:delete", RefreshBranches )
            , ( "build:new", AddBuildEvent )
            , ( "build:delete", DeleteBuildEvent )
            , ( "build:update", UpdateBuildEvent )
            ]

        merge e =
            let
                existsInBoth key a b dict =
                    Dict.insert key (List.append a b) dict
            in
                Dict.merge Dict.insert existsInBoth Dict.insert Dict.empty subPageEvents e
    in
        Dict.singleton (channelName slug) pageEvents
            |> merge


loadedEvents : Msg -> Model -> Dict String (List ( String, Encode.Value -> Msg ))
loadedEvents msg model =
    case ( msg, getSubPage model.subPageState ) of
        ( CommitMsg subMsg, Commit subModel ) ->
            Commit.loadedEvents subMsg subModel
                |> Dict.map (\_ v -> List.map (Tuple.mapSecond (\msg -> msg >> CommitMsg)) v)

        _ ->
            Dict.empty


leaveChannels : Model -> Maybe Project.Slug -> Maybe ProjectRoute.Route -> List String
leaveChannels model maybeProjectSlug maybeProjectRoute =
    let
        currentProjectSlug =
            model.project.slug

        projectChannel =
            channelName currentProjectSlug
    in
        case ( getSubPage model.subPageState, maybeProjectSlug, maybeProjectRoute ) of
            ( Commit subModel, Just routeProjectSlug, Just (ProjectRoute.Commit _ commitRoute) ) ->
                if routeProjectSlug == currentProjectSlug then
                    Commit.leaveChannels subModel (Just commitRoute)
                else
                    projectChannel :: Commit.leaveChannels subModel Nothing

            ( Commit subModel, Just routeProjectSlug, _ ) ->
                if routeProjectSlug == currentProjectSlug then
                    Commit.leaveChannels subModel Nothing
                else
                    projectChannel :: Commit.leaveChannels subModel Nothing

            ( Commit subModel, _, _ ) ->
                projectChannel :: Commit.leaveChannels subModel Nothing

            ( _, Just routeProjectSlug, _ ) ->
                if routeProjectSlug == currentProjectSlug then
                    []
                else
                    [ projectChannel ]

            _ ->
                [ projectChannel ]



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    case (getSubPage model.subPageState) of
        Commits subModel ->
            Commits.subscriptions model.branches subModel
                |> Sub.map CommitsMsg

        _ ->
            Sub.none



-- VIEW --


view : Session msg -> Model -> Html Msg
view session model =
    viewSubPage session model


sidebarConfig : Sidebar.Config Msg
sidebarConfig =
    { newUrlMsg = NewUrl
    , commitPopMsg = CommitsIconPopMsg
    , settingsPopMsg = SettingsIconPopMsg
    }


viewSubPage : Session msg -> Model -> Html Msg
viewSubPage session model =
    let
        page =
            getSubPage model.subPageState

        project =
            model.project

        branches =
            model.branches

        breadcrumb =
            viewBreadcrumb project

        sidebar =
            Sidebar.view model sidebarConfig project

        pageFrame =
            frame project
    in
        case page of
            Blank ->
                Html.text ""
                    |> pageFrame (sidebar OtherPage) (breadcrumb (text "") [])

            Errored subModel ->
                Html.text "Errored"
                    |> pageFrame (sidebar OtherPage) (breadcrumb (text "") [])

            Overview _ ->
                Overview.view project model.builds
                    |> Html.map OverviewMsg
                    |> pageFrame (sidebar OverviewPage) (breadcrumb (text "") [])

            Commits subModel ->
                let
                    crumb =
                        Commits.breadcrumb project
                            |> breadcrumb crumbExtraItems

                    crumbExtraItems =
                        Commits.viewBreadcrumbExtraItems project subModel
                            |> Html.map CommitsMsg
                in
                    Commits.view project branches subModel
                        |> Html.map CommitsMsg
                        |> pageFrame (sidebar CommitsPage) crumb

            Commit subModel ->
                let
                    crumb =
                        Commit.breadcrumb project subModel.commit subModel.subPageState
                            |> breadcrumb (text "")
                in
                    Commit.view project subModel
                        |> Html.map CommitMsg
                        |> pageFrame (sidebar CommitsPage) crumb

            Settings subModel ->
                let
                    crumb =
                        Settings.breadcrumb project
                            |> breadcrumb (text "")
                in
                    Settings.view project subModel
                        |> Html.map SettingsMsg
                        |> pageFrame (sidebar SettingsPage) crumb


frame : Project -> Html msg -> Html msg -> Html msg -> Html msg
frame project sidebar breadcrumb content =
    div []
        [ sidebar
        , div [ class "project-content-container px-4" ]
            [ breadcrumb
            , content
            ]
        ]



-- BREADCRUMB


viewBreadcrumb : Project -> Html Msg -> List ( Route, String ) -> Html Msg
viewBreadcrumb project additionalElements items =
    let
        fixedItems =
            [ ( Route.Projects, "Projects" )
            , ( Route.Project project.slug ProjectRoute.Overview, project.name )
            ]

        allItems =
            fixedItems ++ items

        allItemLength =
            List.length allItems

        breadcrumbItem i item =
            item |> viewBreadcrumbItem (i == (allItemLength - 1))

        itemElements =
            List.indexedMap breadcrumbItem allItems
    in
        div [ class "row" ]
            [ ol [ class "breadcrumb bg-white" ] itemElements
            , additionalElements
            ]


viewBreadcrumbItem : Bool -> ( Route, String ) -> Html Msg
viewBreadcrumbItem active ( route, name ) =
    let
        children =
            if active then
                text name
            else
                a [ Route.href route, class "text-secondary" ] [ text name ]
    in
        li
            [ Route.href route
            , onClickPage NewUrl route
            , class "breadcrumb-item"
            , classList [ ( "active", active ) ]
            ]
            [ children ]



-- UPDATE --


type Msg
    = NewUrl String
    | SetRoute (Maybe ProjectRoute.Route)
    | CommitsMsg Commits.Msg
    | OverviewMsg Overview.Msg
    | CommitsLoaded (Result PageLoadError Commits.Model)
    | CommitMsg Commit.Msg
    | CommitLoaded (Result PageLoadError ( Commit.Model, Cmd Commit.Msg ))
    | SettingsMsg Settings.Msg
    | UpdateProject Encode.Value
    | AddBranch Encode.Value
    | ProjectDeleted Encode.Value
    | RefreshBranches Encode.Value
    | RefreshBranchesComplete (Result Request.Errors.HttpError (PaginatedList Branch))
    | AddBuildEvent Encode.Value
    | UpdateBuildEvent Encode.Value
    | DeleteBuildEvent Encode.Value
    | CommitsIconPopMsg Popover.State
    | SettingsIconPopMsg Popover.State


getSubPage : SubPageState -> SubPage
getSubPage subPageState =
    case subPageState of
        Loaded subPage ->
            subPage

        TransitioningFrom subPage ->
            subPage


setRoute : Context -> Session msg -> Maybe ProjectRoute.Route -> Model -> ( Model, Cmd Msg )
setRoute context session maybeRoute model =
    let
        transition toMsg task =
            { model | subPageState = TransitioningFrom (getSubPage model.subPageState) }
                => Task.attempt toMsg task

        errored =
            pageErrored model
    in
        case maybeRoute of
            Nothing ->
                { model | subPageState = Loaded Blank } => Cmd.none

            Just (ProjectRoute.Overview) ->
                case session.user of
                    Just user ->
                        { model | subPageState = Loaded (Overview Overview.initialModel) } => Cmd.none

                    Nothing ->
                        errored Page.Project "Uhoh"

            Just (ProjectRoute.Commits maybeBranch maybePage) ->
                case session.user of
                    Just user ->
                        Commits.init context session model.branches model.project.slug maybeBranch maybePage
                            |> transition CommitsLoaded

                    Nothing ->
                        errored Page.Project "Uhoh"

            Just (ProjectRoute.Commit hash maybeRoute) ->
                let
                    loadFreshPage =
                        Just maybeRoute
                            |> Commit.init context session model.project hash
                            |> transition CommitLoaded

                    transitionSubPage subModel =
                        let
                            ( newModel, newMsg ) =
                                subModel
                                    |> Commit.update context model.project session (Commit.SetRoute (Just maybeRoute))
                        in
                            { model | subPageState = Loaded (Commit newModel) }
                                => Cmd.map CommitMsg newMsg
                in
                    case ( session.user, model.subPageState ) of
                        ( Just _, Loaded page ) ->
                            case page of
                                -- If we're on the product page for the same product as the new route just load sub-page
                                -- Otherwise load the project page fresh
                                Commit subModel ->
                                    if hash == subModel.commit.hash then
                                        transitionSubPage subModel
                                    else
                                        loadFreshPage

                                _ ->
                                    loadFreshPage

                        ( Just _, TransitioningFrom _ ) ->
                            loadFreshPage

                        ( Nothing, _ ) ->
                            errored Page.Project "Error loading commit"

            Just (ProjectRoute.Settings) ->
                case session.user of
                    Just user ->
                        { model | subPageState = Loaded (Settings (Settings.initialModel)) } => Cmd.none

                    Nothing ->
                        errored Page.Project "Uhoh"


pageErrored : Model -> ActivePage -> String -> ( Model, Cmd msg )
pageErrored model activePage errorMessage =
    let
        error =
            Errored.pageLoadError activePage errorMessage
    in
        { model | subPageState = Loaded (Errored error) } => Cmd.none


update : Context -> Session msg -> Msg -> Model -> ( Model, Cmd Msg )
update context session msg model =
    updateSubPage context session (getSubPage model.subPageState) msg model


updateSubPage : Context -> Session msg -> SubPage -> Msg -> Model -> ( Model, Cmd Msg )
updateSubPage context session subPage msg model =
    let
        toPage toModel toMsg subUpdate subMsg subModel =
            let
                ( newModel, newCmd ) =
                    subUpdate subMsg subModel
            in
                ( { model | subPageState = Loaded (toModel newModel) }, Cmd.map toMsg newCmd )

        errored =
            pageErrored model
    in
        case ( msg, subPage ) of
            ( NewUrl url, _ ) ->
                model
                    => newUrl url

            ( CommitsIconPopMsg state, _ ) ->
                { model | commitIconPopover = state }
                    => Cmd.none

            ( SettingsIconPopMsg state, _ ) ->
                { model | settingsIconPopover = state }
                    => Cmd.none

            ( SetRoute route, _ ) ->
                setRoute context session route model

            ( SettingsMsg subMsg, Settings subModel ) ->
                toPage Settings SettingsMsg (Settings.update context model.project session) subMsg subModel

            ( CommitsLoaded (Ok subModel), _ ) ->
                { model | subPageState = Loaded (Commits subModel) }
                    => Cmd.none

            ( CommitsLoaded (Err error), _ ) ->
                { model | subPageState = Loaded (Errored error) }
                    => Cmd.none

            ( CommitsMsg subMsg, Commits subModel ) ->
                toPage Commits CommitsMsg (Commits.update context model.project session) subMsg subModel

            ( OverviewMsg subMsg, Overview subModel ) ->
                toPage Overview OverviewMsg (Overview.update model.project session) subMsg subModel

            ( CommitLoaded (Ok ( subModel, subMsg )), _ ) ->
                { model | subPageState = Loaded (Commit subModel) }
                    => Cmd.map CommitMsg subMsg

            ( CommitLoaded (Err error), _ ) ->
                { model | subPageState = Loaded (Errored (Debug.log "ERROR " error)) }
                    => Cmd.none

            ( CommitMsg subMsg, Commit subModel ) ->
                let
                    ( newSubModel, newCmd ) =
                        Commit.update context model.project session subMsg subModel
                in
                    { model | subPageState = Loaded (Commit newSubModel) }
                        ! [ Cmd.map CommitMsg newCmd ]

            ( UpdateProject updateJson, _ ) ->
                let
                    newProject =
                        updateJson
                            |> Decode.decodeValue Project.decoder
                            |> Result.toMaybe
                            |> Maybe.withDefault model.project
                in
                    { model | project = newProject }
                        => Cmd.none

            ( AddBranch branchJson, _ ) ->
                let
                    branches =
                        Decode.decodeValue Branch.decoder branchJson
                            |> Result.toMaybe
                            |> Maybe.map (\b -> b :: model.branches)
                            |> Maybe.withDefault model.branches
                in
                    { model | branches = branches }
                        => Cmd.none

            ( RefreshBranches _, _ ) ->
                model
                    => Task.attempt RefreshBranchesComplete (Request.Project.branches context model.project.slug (Maybe.map .token session.user))

            ( RefreshBranchesComplete (Ok (Paginated { results })), _ ) ->
                { model | branches = results }
                    => Cmd.none

            ( ProjectDeleted _, _ ) ->
                model
                    => Route.modifyUrl Route.Projects

            ( AddBuildEvent buildJson, _ ) ->
                let
                    builds =
                        Decode.decodeValue Build.decoder buildJson
                            |> Result.toMaybe
                            |> Maybe.map (addBuild model.builds)
                            |> Maybe.withDefault model.builds
                in
                    { model | builds = sortByDatetime .createdAt builds }
                        => Cmd.none

            ( DeleteBuildEvent buildJson, _ ) ->
                let
                    builds =
                        Decode.decodeValue Build.decoder buildJson
                            |> Result.toMaybe
                            |> Maybe.map (\b -> List.filter (\a -> b.id /= a.id) model.builds)
                            |> Maybe.withDefault model.builds
                in
                    { model | builds = builds }
                        => Cmd.none

            ( UpdateBuildEvent buildJson, _ ) ->
                let
                    builds =
                        Decode.decodeValue Build.decoder buildJson
                            |> Result.toMaybe
                            |> Maybe.map
                                (\b ->
                                    List.map
                                        (\a ->
                                            if b.id == a.id then
                                                b
                                            else
                                                a
                                        )
                                        model.builds
                                )
                            |> Maybe.withDefault model.builds
                in
                    { model | builds = builds }
                        => Cmd.none

            ( _, _ ) ->
                -- Disregard incoming messages that arrived for the wrong sub page
                (Debug.log "Fell through" model)
                    => Cmd.none
