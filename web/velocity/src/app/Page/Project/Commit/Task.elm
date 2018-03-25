module Page.Project.Commit.Task exposing (..)

import Ansi.Log
import Context exposing (Context)
import Data.Commit as Commit exposing (Commit)
import Data.Project as Project exposing (Project)
import Data.Session as Session exposing (Session)
import Data.Build as Build exposing (Build)
import Data.BuildStep as BuildStep exposing (BuildStep)
import Data.BuildStream as BuildStream exposing (Id, BuildStream, BuildStreamOutput)
import Data.Task as ProjectTask exposing (Step(..), Parameter(..))
import Data.AuthToken as AuthToken exposing (AuthToken)
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput, on, onSubmit)
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Page.Helpers exposing (validClasses, formatDateTime)
import Request.Commit
import Request.Build
import Request.Errors
import Task exposing (Task)
import Util exposing ((=>))
import Validate exposing (..)
import Views.Form as Form
import Views.Page as Page
import Json.Decode as Decode
import Html.Events.Extra exposing (targetSelectedIndex)
import Route
import Page.Project.Route as ProjectRoute
import Page.Project.Commit.Route as CommitRoute
import Views.Task exposing (viewStepList)
import Views.Helpers exposing (onClickPage)
import Navigation
import Dict exposing (Dict)
import Json.Encode as Encode
import Array exposing (Array)
import Html.Lazy as Lazy
import Views.Build exposing (viewBuildStatusIcon, viewBuildStepStatusIcon, viewBuildTextClass)
import Component.BuildOutput as BuildOutput


-- MODEL --


type alias Model =
    { task : ProjectTask.Task
    , toggledStep : Maybe Step
    , form : List Field
    , errors : List Error
    , selectedTab : Tab
    , frame : Frame
    }


type alias InputFormField =
    { value : String
    , dirty : Bool
    , field : String
    }


type alias ChoiceFormField =
    { value : Maybe String
    , dirty : Bool
    , field : String
    , options : List String
    }


type Field
    = Input InputFormField
    | Choice ChoiceFormField


type alias FromBuild =
    Build.Id


type alias ToBuild =
    Build.Id


type Stream
    = Stream BuildStream.Id


type Tab
    = NewFormTab
    | BuildTab Int


type Frame
    = BuildFrame BuildType
    | NewFormFrame


type BuildType
    = LoadedBuild Build.Id BuildOutput.Model
    | LoadingBuild (Maybe FromBuild) (Maybe ToBuild)


buildOutput : Array BuildStreamOutput -> Ansi.Log.Model
buildOutput buildOutput =
    Array.foldl Ansi.Log.update
        (Ansi.Log.init Ansi.Log.Cooked)
        (Array.map .output buildOutput)


stringToTab : Maybe String -> List Build -> Tab
stringToTab maybeSelectedTab builds =
    case maybeSelectedTab of
        Just "new" ->
            NewFormTab

        Just tabText ->
            tabText
                |> String.split "-"
                |> List.reverse
                |> List.head
                |> Maybe.andThen (String.toInt >> Result.toMaybe)
                |> Maybe.map BuildTab
                |> Maybe.withDefault NewFormTab

        Nothing ->
            NewFormTab


init : Context -> Session msg -> Project.Id -> Commit.Hash -> ProjectTask.Task -> Maybe String -> List Build -> Task PageLoadError Model
init context session id hash task maybeSelectedTab builds =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        handleLoadError _ =
            pageLoadError Page.Project "Project unavailable."

        selectedTab =
            stringToTab maybeSelectedTab builds

        initialModel frame =
            let
                toggledStep =
                    Nothing

                form =
                    List.filterMap newField task.parameters

                errors =
                    List.concatMap validator form
            in
                { task = task
                , toggledStep = toggledStep
                , form = form
                , errors = errors
                , selectedTab = selectedTab
                , frame = Maybe.withDefault NewFormFrame frame
                }
    in
        case selectedTab of
            NewFormTab ->
                Task.succeed (initialModel (Just NewFormFrame))

            BuildTab buildIndex ->
                let
                    build =
                        builds
                            |> Array.fromList
                            |> Array.get (buildIndex - 1)
                in
                    case build of
                        Just b ->
                            BuildOutput.init context task maybeAuthToken b
                                |> Task.map (LoadedBuild b.id >> BuildFrame >> Just >> initialModel)
                                |> Task.mapError handleLoadError

                        Nothing ->
                            Task.succeed (initialModel (Just NewFormFrame))


newField : Parameter -> Maybe Field
newField parameter =
    case parameter of
        StringParam param ->
            let
                value =
                    Maybe.withDefault "" param.default

                dirty =
                    String.length value > 0
            in
                InputFormField value dirty param.name
                    |> Input
                    |> Just

        ChoiceParam param ->
            let
                options =
                    param.default
                        :: (List.map Just param.options)
                        |> List.filterMap identity

                value =
                    case param.default of
                        Nothing ->
                            List.head options

                        default ->
                            default
            in
                ChoiceFormField value True param.name options
                    |> Choice
                    |> Just

        DerivedParam _ ->
            Nothing



-- CHANNELS --


streamChannelName : BuildStream -> String
streamChannelName stream =
    "stream:" ++ (BuildStream.idToString stream.id)


buildEvents : List Build -> Build.Id -> Dict String (List ( String, Encode.Value -> Msg ))
buildEvents builds buildId =
    let
        build =
            builds
                |> List.filter (\b -> b.id == buildId)
                |> List.head

        streams =
            build
                |> Maybe.map (\build -> List.map .streams build.steps)
                |> Maybe.map (List.foldr (++) [])
                |> Maybe.withDefault []

        foldStreamEvents stream dict =
            let
                channelName =
                    streamChannelName stream

                events =
                    [ ( "streamLine:new", AddStreamOutput stream ) ]
            in
                Dict.insert channelName events dict
    in
        List.foldl foldStreamEvents Dict.empty streams


events : List Build -> Model -> Dict String (List ( String, Encode.Value -> Msg ))
events builds model =
    case model.frame of
        BuildFrame (LoadedBuild id _) ->
            buildEvents builds id

        _ ->
            Dict.empty


leaveChannels : List Build -> Model -> Maybe String -> List String
leaveChannels builds model maybeBuildId =
    let
        channels id b =
            if id == Build.idToString b then
                []
            else
                Dict.keys (buildEvents builds b)
    in
        case ( maybeBuildId, model.frame ) of
            ( Just buildId, BuildFrame (LoadedBuild id _) ) ->
                channels buildId id

            ( Just buildId, BuildFrame (LoadingBuild (Just id) _) ) ->
                channels buildId id

            ( _, BuildFrame (LoadedBuild id _) ) ->
                Dict.keys (buildEvents builds id)

            ( _, BuildFrame (LoadingBuild (Just b) _) ) ->
                Dict.keys (buildEvents builds b)

            _ ->
                []



-- VIEW --


view : Project -> Commit -> Model -> List Build -> Html Msg
view project commit model builds =
    let
        task =
            model.task

        stepList =
            viewStepList task.steps model.toggledStep

        navigation =
            if List.isEmpty builds then
                text ""
            else
                viewTabs project commit task builds model.selectedTab
    in
        div [ class "row" ]
            [ div [ class "col-sm-12 col-md-12 col-lg-12 default-margin-bottom" ]
                [ viewTaskDescriptor task
                , navigation
                , Lazy.lazy (viewTabFrame model) builds
                ]
            ]


viewTaskDescriptor : ProjectTask.Task -> Html Msg
viewTaskDescriptor task =
    div [ class "card mb-3 border-secondary" ]
        [ div [ class "card-body" ]
            [ h3 []
                [ text (ProjectTask.nameToString task.name)
                , text " "
                , small [ class "text-muted" ] [ text task.description ]
                ]
            ]
        ]


viewTabs : Project -> Commit -> ProjectTask.Task -> List Build -> Tab -> Html Msg
viewTabs project commit task builds selectedTab =
    let
        compare a b =
            case ( a, b ) of
                ( BuildTab c, BuildTab d ) ->
                    c == d

                ( NewFormTab, NewFormTab ) ->
                    True

                _ ->
                    False

        buildTab t =
            let
                build =
                    case t of
                        NewFormTab ->
                            Nothing

                        BuildTab b ->
                            Array.fromList builds
                                |> Array.get (b - 1)

                tabContent =
                    case t of
                        NewFormTab ->
                            i [ class "fa fa-plus-circle text-secondary" ] []

                        BuildTab b ->
                            if List.length builds == 1 then
                                text "Build output "
                            else
                                text ("Build output #" ++ (toString b) ++ " ")

                tabQueryParam =
                    case t of
                        NewFormTab ->
                            "new"

                        BuildTab b ->
                            "build-" ++ (toString b)

                route =
                    Just tabQueryParam
                        |> CommitRoute.Task task.name
                        |> ProjectRoute.Commit commit.hash
                        |> Route.Project project.slug

                tabIcon =
                    build
                        |> Maybe.map viewBuildStatusIcon
                        |> Maybe.withDefault (text "")

                textClass =
                    build
                        |> Maybe.map viewBuildTextClass
                        |> Maybe.withDefault ("")

                tabClassList =
                    [ ( "nav-link", True )
                    , ( textClass, True )
                    , ( "active", compare t selectedTab )
                    ]
            in
                li [ class "nav-item" ]
                    [ a
                        [ classList tabClassList
                        , Route.href route
                        , onClickPage (SelectTab selectedTab) route
                        ]
                        [ tabContent
                        , text " "
                        , tabIcon
                        ]
                    ]

        buildTabs =
            builds
                |> List.indexedMap (\i -> (\_ -> buildTab (BuildTab (i + 1))))
    in
        List.append buildTabs [ buildTab NewFormTab ]
            |> ul [ class "nav nav-tabs nav-fill" ]


viewTabFrame : Model -> List Build -> Html Msg
viewTabFrame model builds =
    let
        buildForm =
            div [] <|
                viewBuildForm (ProjectTask.nameToString model.task.name) model.form model.errors

        findBuild id =
            builds
                |> List.filter (\a -> a.id == id)
                |> List.head
    in
        case model.frame of
            NewFormFrame ->
                buildForm

            BuildFrame (LoadedBuild _ buildOutputModel) ->
                BuildOutput.view buildOutputModel

            _ ->
                text ""


viewBuildForm : String -> List Field -> List Error -> List (Html Msg)
viewBuildForm taskName fields errors =
    let
        fieldInput f =
            case f of
                Choice field ->
                    let
                        value =
                            Maybe.withDefault "" field.value

                        option o =
                            Html.option
                                [ selected (o == value) ]
                                [ text o ]
                    in
                        Form.select
                            { name = field.field
                            , label = field.field
                            , help = Nothing
                            , errors = []
                            }
                            [ attribute "required" ""
                            , classList (validClasses errors field)
                            , on "change" <| Decode.map (OnChange field) targetSelectedIndex
                            ]
                            (List.map option field.options)

                Input field ->
                    Form.input
                        { name = field.field
                        , label = field.field
                        , help = Nothing
                        , errors = []
                        }
                        [ attribute "required" ""
                        , value field.value
                        , onInput (OnInput field)
                        , classList (validClasses errors field)
                        ]
                        []
    in
        [ Html.form [ class "mt-3", attribute "novalidate" "", onSubmit SubmitForm ] <|
            List.map fieldInput fields
                ++ [ button
                        [ class "btn btn-lg btn-block btn-primary"
                        , type_ "submit"
                        , disabled <| not (List.isEmpty errors)
                        ]
                        [ text "Start task" ]
                   ]
        ]


breadcrumb : Project -> Commit -> ProjectTask.Task -> List ( Route.Route, String )
breadcrumb project commit task =
    [ ( CommitRoute.Task task.name Nothing |> ProjectRoute.Commit commit.hash |> Route.Project project.slug
      , ProjectTask.nameToString task.name
      )
    ]



-- UPDATE --


type Msg
    = ToggleStep (Maybe Step)
    | OnInput InputFormField String
    | OnChange ChoiceFormField (Maybe Int)
    | SubmitForm
    | BuildCreated (Result Request.Errors.HttpError Build)
    | SelectTab Tab String
    | LoadBuild Build
    | BuildLoaded (Result Request.Errors.HttpError (Maybe BuildType))
    | AddStreamOutput BuildStream Encode.Value
    | BuildUpdated Encode.Value


type ExternalMsg
    = NoOp
    | AddBuild Build
    | UpdateBuild Build


update : Context -> Project -> Commit -> List Build -> Session msg -> Msg -> Model -> ( ( Model, Cmd Msg ), ExternalMsg )
update context project commit builds session msg model =
    let
        projectSlug =
            project.slug

        commitHash =
            commit.hash

        taskName =
            model.task.name

        maybeAuthToken =
            Maybe.map .token session.user
    in
        case msg of
            ToggleStep maybeStep ->
                { model | toggledStep = maybeStep }
                    => Cmd.none
                    => NoOp

            OnInput field value ->
                let
                    updateField fieldType =
                        case fieldType of
                            Input f ->
                                if f == field then
                                    Input
                                        { field
                                            | value = value
                                            , dirty = True
                                        }
                                else
                                    fieldType

                            _ ->
                                fieldType

                    form =
                        List.map updateField model.form

                    errors =
                        List.concatMap validator form
                in
                    { model
                        | form = form
                        , errors = errors
                    }
                        => Cmd.none
                        => NoOp

            OnChange field maybeIndex ->
                let
                    updateField fieldType =
                        case ( fieldType, maybeIndex ) of
                            ( Choice f, Just index ) ->
                                if f == field then
                                    let
                                        value =
                                            f.options
                                                |> List.indexedMap (,)
                                                |> List.filter (\( i, _ ) -> i == index)
                                                |> List.head
                                                |> Maybe.map Tuple.second
                                    in
                                        Choice
                                            { field
                                                | value = value
                                                , dirty = True
                                            }
                                else
                                    fieldType

                            _ ->
                                fieldType

                    form =
                        List.map updateField model.form

                    errors =
                        List.concatMap validator form
                in
                    { model
                        | form = form
                        , errors = errors
                    }
                        => Cmd.none
                        => NoOp

            SubmitForm ->
                let
                    stringParam { value, field } =
                        field => value

                    cmdFromAuth authToken =
                        authToken
                            |> Request.Commit.createBuild context projectSlug commitHash taskName params
                            |> Task.attempt BuildCreated

                    cmd =
                        session
                            |> Session.attempt "create build" cmdFromAuth
                            |> Tuple.second

                    mapFieldToParam field =
                        case field of
                            Input input ->
                                Just (stringParam input)

                            Choice choice ->
                                choice.value
                                    |> Maybe.map (\value -> stringParam { value = value, field = choice.field })

                    params =
                        List.filterMap mapFieldToParam model.form
                in
                    model
                        => cmd
                        => NoOp

            LoadBuild build ->
                model => Cmd.none => NoOp

            --                let
            --                    cmd =
            --                        build
            --                            |> BuildOutput.init context model.task maybeAuthToken
            --                            |> Task.attempt BuildLoaded
            --                in
            --                    model
            --                        => cmd
            --                        => NoOp
            BuildLoaded (Ok (Just loadedBuild)) ->
                model
                    => Cmd.none
                    => NoOp

            BuildLoaded _ ->
                model
                    => Cmd.none
                    => NoOp

            BuildCreated (Ok build) ->
                let
                    tabNum =
                        List.length builds

                    tab =
                        tabNum
                            |> toString
                            |> (\i -> "build-" ++ i)
                            |> Just

                    route =
                        CommitRoute.Task model.task.name tab
                            |> ProjectRoute.Commit commit.hash
                            |> Route.Project project.slug
                in
                    model
                        => Navigation.newUrl (Route.routeToString route)
                        => AddBuild build

            BuildUpdated buildJson ->
                let
                    externalMsg =
                        buildJson
                            |> Decode.decodeValue Build.decoder
                            |> Result.toMaybe
                            |> Maybe.map UpdateBuild
                            |> Maybe.withDefault NoOp
                in
                    model
                        => Cmd.none
                        => externalMsg

            BuildCreated (Err _) ->
                model
                    => Cmd.none
                    => NoOp

            SelectTab tab url ->
                let
                    frame =
                        case tab of
                            BuildTab toBuildIndex ->
                                let
                                    toBuild =
                                        builds
                                            |> Array.fromList
                                            |> Array.get (toBuildIndex - 1)
                                            |> Maybe.map .id

                                    fromBuild =
                                        case model.frame of
                                            BuildFrame (LoadedBuild b _) ->
                                                Just b

                                            _ ->
                                                Nothing
                                in
                                    BuildFrame (LoadingBuild fromBuild toBuild)

                            NewFormTab ->
                                NewFormFrame
                in
                    { model
                        | selectedTab = tab
                        , frame = frame
                    }
                        => Navigation.newUrl url
                        => NoOp

            AddStreamOutput buildStream outputJson ->
                model => Cmd.none => NoOp



--                let
--                    frame =
--                        case model.frame of
--                            BuildFrame (LoadedBuild build streams) ->
--                                outputJson
--                                    |> Decode.decodeValue BuildStream.outputDecoder
--                                    |> Result.toMaybe
--                                    |> Maybe.map
--                                        (\b ->
--                                            let
--                                                streamKey =
--                                                    BuildStream.idToString buildStream.id
--
--                                                streamLines =
--                                                    Dict.get streamKey streams
--                                            in
--                                                case streamLines of
--                                                    Just ( number, taskStep, buildStepId, streamLines, ansi ) ->
--                                                        let
--                                                            streamLineLength =
--                                                                Array.length streamLines - 1
--
--                                                            updatedStreamLines =
--                                                                if b.line > streamLineLength then
--                                                                    Array.push b streamLines
--                                                                else
--                                                                    Array.set b.line b streamLines
--
--                                                            lineAnsi outputLine ansi =
--                                                                Ansi.Log.update outputLine.output ansi
--
--                                                            updatedAnsi =
--                                                                Array.foldl lineAnsi ansi (Array.initialize 1 (always b))
--
--                                                            streamTuple =
--                                                                ( number, taskStep, buildStepId, updatedStreamLines, updatedAnsi )
--                                                        in
--                                                            Dict.insert streamKey streamTuple streams
--
--                                                    _ ->
--                                                        streams
--                                        )
--                                    |> Maybe.withDefault streams
--                                    |> LoadedBuild build
--                                    |> BuildFrame
--
--                            _ ->
--                                model.frame
--                in
--                    { model | frame = frame }
--                        => Cmd.none
--                        => NoOp
-- VALIDATION --


type alias Error =
    ( String, String )


validator : Validator Error Field
validator =
    [ \f ->
        let
            notBlank { field, value } =
                ifBlank (field => "Field cannot be blank") value
        in
            case f of
                Input fieldType ->
                    notBlank fieldType

                Choice fieldType ->
                    (\{ field, value } ->
                        value
                            |> Maybe.withDefault ""
                            |> ifBlank (field => "Field cannot be blank")
                    )
                        fieldType
    ]
        |> Validate.all
