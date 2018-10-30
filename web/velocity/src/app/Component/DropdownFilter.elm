module Component.DropdownFilter exposing (Config, Context, DropdownState, initialDropdownState, subscriptions, view)

{- A stateless ItemFilter component. -}
-- INTERNAL --
-- EXTERNAL --

import Bootstrap.Button as Button
import Bootstrap.Dropdown as Dropdown
import Bootstrap.Form as Form
import Bootstrap.Form.Input as Input
import Data.Device as Device
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onInput, onWithOptions)
import Json.Decode as Decode


-- MODEL --


type alias Context a =
    { items : List a
    , dropdownState : Dropdown.State
    , filterTerm : String
    , selectedItem : Maybe a
    , deviceSize : Device.Size
    }


type alias Config msg item =
    { dropdownMsg : Dropdown.State -> msg
    , termMsg : String -> msg
    , noOpMsg : msg
    , selectItemMsg : Maybe item -> msg
    , labelFn : item -> String
    , icon : Html msg
    , showFilter : Bool
    , showAllItemsItem : Bool
    }


type alias DropdownState =
    Dropdown.State


initialDropdownState : DropdownState
initialDropdownState =
    Dropdown.initialState



-- SUBSCRIPTIONS --


subscriptions : Config msg a -> Context a -> Sub msg
subscriptions { dropdownMsg } { dropdownState } =
    Dropdown.subscriptions dropdownState dropdownMsg



-- VIEW --


view : Config msg a -> Context a -> Html msg
view config context =
    if Device.isSmall context.deviceSize then
        select
            [ class "custom-select custom-select-lg"
            , onInput
                (\itemLabel ->
                    List.filter (\item -> config.labelFn item == itemLabel) context.items
                        |> List.head
                        |> config.selectItemMsg
                )
            ]
            (viewChoiceItems context config)
    else
        Dropdown.dropdown
            context.dropdownState
            { options =
                [ Dropdown.menuAttrs
                    [ onClick config.noOpMsg
                    , class "item-filter-dropdown"
                    ]
                ]
            , toggleMsg = config.dropdownMsg
            , toggleButton = toggleButton context config
            , items = viewDropdownItems config context
            }


toggleButton : Context a -> Config msg a -> Dropdown.DropdownToggle msg
toggleButton { selectedItem } config =
    let
        toggleText =
            itemLabelString config selectedItem
    in
        Dropdown.toggle
            [ Button.outlineSecondary
            ]
            [ config.icon
            , text (" " ++ toggleText)
            ]


viewChoiceItems : Context a -> Config msg a -> List (Html msg)
viewChoiceItems { items, selectedItem } { labelFn } =
    List.map
        (\item ->
            option
                [ selected (selectedItem == Just item)
                , value (labelFn item)
                ]
                [ text (labelFn item) ]
        )
        items


viewDropdownItems : Config msg a -> Context a -> List (Dropdown.DropdownItem msg)
viewDropdownItems config context =
    let
        filter existing =
            if config.showFilter then
                filterForm config context :: existing
            else
                existing

        noItemSelectedItems existing =
            if config.showAllItemsItem then
                [ Dropdown.divider
                , noItemSelected config context
                , Dropdown.divider
                ]
                    ++ existing
            else
                existing
    in
        viewItems config context
            |> noItemSelectedItems
            |> filter


filterForm : Config msg a -> Context a -> Dropdown.DropdownItem msg
filterForm config context =
    Dropdown.customItem (viewForm config.termMsg context)


noItemSelected : Config msg a -> Context a -> Dropdown.DropdownItem msg
noItemSelected config { selectedItem } =
    viewItem config selectedItem Nothing


viewItems : Config msg a -> Context a -> List (Dropdown.DropdownItem msg)
viewItems config { items, filterTerm, selectedItem } =
    items
        |> List.filter (config.labelFn >> String.contains filterTerm)
        |> List.map (Just >> viewItem config selectedItem)


itemLabelString : Config msg a -> Maybe a -> String
itemLabelString config maybeItem =
    maybeItem
        |> Maybe.map config.labelFn
        |> Maybe.withDefault "all-items"


viewItem : Config msg a -> Maybe a -> Maybe a -> Dropdown.DropdownItem msg
viewItem config selectedItem maybeItem =
    let
        itemIcon =
            if selectedItem == maybeItem then
                i [ class "fa fa-check" ] []
            else
                text ""
    in
        Dropdown.buttonItem
            [ onClick (config.selectItemMsg <| maybeItem) ]
            [ itemIcon
            , text (itemLabelString config maybeItem)
            ]


viewForm : (String -> msg) -> Context a -> Html msg
viewForm msg { filterTerm } =
    Form.form [ class "px-2 py-0 filter-list-select", style "width" "400px" ]
        [ Form.group []
            [ Input.text
                [ Input.placeholder "Filter items"
                , Input.attrs [ onInput msg, id "filter-item-input" ]
                , Input.value filterTerm
                ]
            ]
        ]



-- helper to cancel click anywhere


onClick : msg -> Attribute msg
onClick message =
    onWithOptions
        "click"
        { stopPropagation = True
        , preventDefault = False
        }
        (Decode.succeed message)
