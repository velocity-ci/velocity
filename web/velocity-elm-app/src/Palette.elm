module Palette exposing (black, danger1, danger2, danger3, danger4, danger5, danger6, danger7, info1, info2, info3, info4, info5, info6, info7, neutral1, neutral2, neutral3, neutral4, neutral5, neutral6, neutral7, primary1, primary2, primary3, primary4, primary5, primary6, primary7, success1, success2, success3, success4, success5, success6, success7, transparent, warning1, warning2, warning3, warning4, warning5, warning6, warning7, white)

import Element exposing (..)


-- Utils


white : Element.Color
white =
    rgba 245 245 245 1


black : Element.Color
black =
    rgba 0 0 0 1


transparent : Element.Color
transparent =
    rgba 0 0 0 0



-- Primary  palette


primary1 : Element.Color
primary1 =
    rgba255 24 54 76 1


primary2 : Element.Color
primary2 =
    rgba255 8 64 105 1


primary3 : Element.Color
primary3 =
    rgba255 0 93 156 1


primary4 : Element.Color
primary4 =
    rgba255 0 120 198 1


primary5 : Element.Color
primary5 =
    rgba255 71 151 215 1


primary6 : Element.Color
primary6 =
    rgba255 150 207 248 1


primary7 : Element.Color
primary7 =
    rgba255 234 247 255 1



-- Neutral palette


neutral1 : Element.Color
neutral1 =
    rgba255 29 37 47 1


neutral2 : Element.Color
neutral2 =
    rgba255 83 96 112 1


neutral3 : Element.Color
neutral3 =
    rgba255 123 138 159 1


neutral4 : Element.Color
neutral4 =
    rgba255 172 189 200 1


neutral5 : Element.Color
neutral5 =
    rgba255 198 208 218 1


neutral6 : Element.Color
neutral6 =
    rgba255 220 227 232 1


neutral7 : Element.Color
neutral7 =
    rgba255 247 248 249 1



-- Accent palette


info1 : Element.Color
info1 =
    rgba255 0 62 60 1


info2 =
    rgba255 0 92 84 1


info3 =
    rgba255 0 136 124 1


info4 =
    rgba255 0 167 153 1


info5 =
    rgba255 44 212 206 1


info6 =
    rgba255 136 237 233 1


info7 =
    rgba255 221 255 254 1


success1 =
    rgba255 0 74 48 1


success2 =
    rgba255 0 111 51 1


success3 =
    rgba255 0 149 69 1


success4 =
    rgba255 0 188 92 1


success5 =
    rgba255 50 216 142 1


success6 =
    rgba255 135 238 181 1


success7 =
    rgba255 215 252 232 1


warning1 =
    rgba255 84 63 5 1


warning2 =
    rgba255 134 96 0 1


warning3 =
    rgba255 201 153 13 1


warning4 =
    rgba255 250 193 63 1


warning5 =
    rgba255 253 222 137 1


warning6 =
    rgba255 255 241 205 1


warning7 =
    rgba255 255 251 242 1


danger1 =
    rgba255 93 16 20 1


danger2 =
    rgba255 136 5 19 1


danger3 =
    rgba255 189 0 14 1


danger4 =
    rgba255 232 0 26 1


danger5 =
    rgba255 240 77 83 1


danger6 =
    rgba255 255 155 159 1


danger7 =
    rgba255 255 228 227 1
