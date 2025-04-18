package utils

// CountryNameMap 英文国家名称到中文国家名称的映射
var CountryNameMap = map[string]string{
	"Afghanistan":                      "阿富汗",
	"Albania":                          "阿尔巴尼亚",
	"Algeria":                          "阿尔及利亚",
	"Andorra":                          "安道尔",
	"Angola":                           "安哥拉",
	"Antigua and Barbuda":              "安提瓜和巴布达",
	"Argentina":                        "阿根廷",
	"Armenia":                          "亚美尼亚",
	"Australia":                        "澳大利亚",
	"Austria":                          "奥地利",
	"Azerbaijan":                       "阿塞拜疆",
	"Bahamas":                          "巴哈马",
	"Bahrain":                          "巴林",
	"Bangladesh":                       "孟加拉国",
	"Barbados":                         "巴巴多斯",
	"Belarus":                          "白俄罗斯",
	"Belgium":                          "比利时",
	"Belize":                           "伯利兹",
	"Benin":                            "贝宁",
	"Bhutan":                           "不丹",
	"Bolivia":                          "玻利维亚",
	"Bosnia and Herzegovina":           "波斯尼亚和黑塞哥维那",
	"Botswana":                         "博茨瓦纳",
	"Brazil":                           "巴西",
	"Brunei":                           "文莱",
	"Bulgaria":                         "保加利亚",
	"Burkina Faso":                     "布基纳法索",
	"Burundi":                          "布隆迪",
	"Cabo Verde":                       "佛得角",
	"Cambodia":                         "柬埔寨",
	"Cameroon":                         "喀麦隆",
	"Canada":                           "加拿大",
	"Central African Republic":         "中非共和国",
	"Chad":                             "乍得",
	"Chile":                            "智利",
	"China":                            "中国",
	"Colombia":                         "哥伦比亚",
	"Comoros":                          "科摩罗",
	"Congo":                            "刚果",
	"Costa Rica":                       "哥斯达黎加",
	"Croatia":                          "克罗地亚",
	"Cuba":                             "古巴",
	"Cyprus":                           "塞浦路斯",
	"Czech Republic":                   "捷克共和国",
	"Denmark":                          "丹麦",
	"Djibouti":                         "吉布提",
	"Dominica":                         "多米尼克",
	"Dominican Republic":               "多米尼加共和国",
	"Ecuador":                          "厄瓜多尔",
	"Egypt":                            "埃及",
	"El Salvador":                      "萨尔瓦多",
	"Equatorial Guinea":                "赤道几内亚",
	"Eritrea":                          "厄立特里亚",
	"Estonia":                          "爱沙尼亚",
	"Eswatini":                         "斯威士兰",
	"Ethiopia":                         "埃塞俄比亚",
	"Fiji":                             "斐济",
	"Finland":                          "芬兰",
	"France":                           "法国",
	"Gabon":                            "加蓬",
	"Gambia":                           "冈比亚",
	"Georgia":                          "格鲁吉亚",
	"Germany":                          "德国",
	"Ghana":                            "加纳",
	"Greece":                           "希腊",
	"Grenada":                          "格林纳达",
	"Guatemala":                        "危地马拉",
	"Guinea":                           "几内亚",
	"Guinea-Bissau":                    "几内亚比绍",
	"Guyana":                           "圭亚那",
	"Haiti":                            "海地",
	"Honduras":                         "洪都拉斯",
	"Hungary":                          "匈牙利",
	"Iceland":                          "冰岛",
	"India":                            "印度",
	"Indonesia":                        "印度尼西亚",
	"Iran":                             "伊朗",
	"Iraq":                             "伊拉克",
	"Ireland":                          "爱尔兰",
	"Israel":                           "以色列",
	"Italy":                            "意大利",
	"Jamaica":                          "牙买加",
	"Japan":                            "日本",
	"Jordan":                           "约旦",
	"Kazakhstan":                       "哈萨克斯坦",
	"Kenya":                            "肯尼亚",
	"Kiribati":                         "基里巴斯",
	"Korea":                            "韩国",
	"South Korea":                      "韩国",
	"North Korea":                      "朝鲜",
	"Kosovo":                           "科索沃",
	"Kuwait":                           "科威特",
	"Kyrgyzstan":                       "吉尔吉斯斯坦",
	"Laos":                             "老挝",
	"Latvia":                           "拉脱维亚",
	"Lebanon":                          "黎巴嫩",
	"Lesotho":                          "莱索托",
	"Liberia":                          "利比里亚",
	"Libya":                            "利比亚",
	"Liechtenstein":                    "列支敦士登",
	"Lithuania":                        "立陶宛",
	"Luxembourg":                       "卢森堡",
	"Madagascar":                       "马达加斯加",
	"Malawi":                           "马拉维",
	"Malaysia":                         "马来西亚",
	"Maldives":                         "马尔代夫",
	"Mali":                             "马里",
	"Malta":                            "马耳他",
	"Marshall Islands":                 "马绍尔群岛",
	"Mauritania":                       "毛里塔尼亚",
	"Mauritius":                        "毛里求斯",
	"Mexico":                           "墨西哥",
	"Micronesia":                       "密克罗尼西亚",
	"Moldova":                          "摩尔多瓦",
	"Monaco":                           "摩纳哥",
	"Mongolia":                         "蒙古",
	"Montenegro":                       "黑山",
	"Morocco":                          "摩洛哥",
	"Mozambique":                       "莫桑比克",
	"Myanmar":                          "缅甸",
	"Namibia":                          "纳米比亚",
	"Nauru":                            "瑙鲁",
	"Nepal":                            "尼泊尔",
	"Netherlands":                      "荷兰",
	"New Zealand":                      "新西兰",
	"Nicaragua":                        "尼加拉瓜",
	"Niger":                            "尼日尔",
	"Nigeria":                          "尼日利亚",
	"North Macedonia":                  "北马其顿",
	"Norway":                           "挪威",
	"Oman":                             "阿曼",
	"Pakistan":                         "巴基斯坦",
	"Palau":                            "帕劳",
	"Palestine":                        "巴勒斯坦",
	"Panama":                           "巴拿马",
	"Papua New Guinea":                 "巴布亚新几内亚",
	"Paraguay":                         "巴拉圭",
	"Peru":                             "秘鲁",
	"Philippines":                      "菲律宾",
	"Poland":                           "波兰",
	"Portugal":                         "葡萄牙",
	"Qatar":                            "卡塔尔",
	"Romania":                          "罗马尼亚",
	"Russia":                           "俄罗斯",
	"Rwanda":                           "卢旺达",
	"Saint Kitts and Nevis":            "圣基茨和尼维斯",
	"Saint Lucia":                      "圣卢西亚",
	"Saint Vincent and the Grenadines": "圣文森特和格林纳丁斯",
	"Samoa":                            "萨摩亚",
	"San Marino":                       "圣马力诺",
	"Sao Tome and Principe":            "圣多美和普林西比",
	"Saudi Arabia":                     "沙特阿拉伯",
	"Senegal":                          "塞内加尔",
	"Serbia":                           "塞尔维亚",
	"Seychelles":                       "塞舌尔",
	"Sierra Leone":                     "塞拉利昂",
	"Singapore":                        "新加坡",
	"Slovakia":                         "斯洛伐克",
	"Slovenia":                         "斯洛文尼亚",
	"Solomon Islands":                  "所罗门群岛",
	"Somalia":                          "索马里",
	"South Africa":                     "南非",
	"South Sudan":                      "南苏丹",
	"Spain":                            "西班牙",
	"Sri Lanka":                        "斯里兰卡",
	"Sudan":                            "苏丹",
	"Suriname":                         "苏里南",
	"Sweden":                           "瑞典",
	"Switzerland":                      "瑞士",
	"Syria":                            "叙利亚",
	"Taiwan":                           "台湾",
	"Tajikistan":                       "塔吉克斯坦",
	"Tanzania":                         "坦桑尼亚",
	"Thailand":                         "泰国",
	"Timor-Leste":                      "东帝汶",
	"Togo":                             "多哥",
	"Tonga":                            "汤加",
	"Trinidad and Tobago":              "特立尼达和多巴哥",
	"Tunisia":                          "突尼斯",
	"Turkey":                           "土耳其",
	"Turkmenistan":                     "土库曼斯坦",
	"Tuvalu":                           "图瓦卢",
	"Uganda":                           "乌干达",
	"Ukraine":                          "乌克兰",
	"United Arab Emirates":             "阿联酋",
	"United Kingdom":                   "英国",
	"United States":                    "美国",
	"USA":                              "美国",
	"Uruguay":                          "乌拉圭",
	"Uzbekistan":                       "乌兹别克斯坦",
	"Vanuatu":                          "瓦努阿图",
	"Vatican City":                     "梵蒂冈",
	"Venezuela":                        "委内瑞拉",
	"Vietnam":                          "越南",
	"Yemen":                            "也门",
	"Zambia":                           "赞比亚",
	"Zimbabwe":                         "津巴布韦",
	"Hong Kong":                        "香港",
	"Macao":                            "澳门",
	"Unknown":                          "未知",
}

// CountryCodeMap 国家代码到中文名称的映射
var CountryCodeMap = map[string]string{
	"AF": "阿富汗",
	"AL": "阿尔巴尼亚",
	"DZ": "阿尔及利亚",
	"AD": "安道尔",
	"AO": "安哥拉",
	"AG": "安提瓜和巴布达",
	"AR": "阿根廷",
	"AM": "亚美尼亚",
	"AU": "澳大利亚",
	"AT": "奥地利",
	"AZ": "阿塞拜疆",
	"BS": "巴哈马",
	"BH": "巴林",
	"BD": "孟加拉国",
	"BB": "巴巴多斯",
	"BY": "白俄罗斯",
	"BE": "比利时",
	"BZ": "伯利兹",
	"BJ": "贝宁",
	"BT": "不丹",
	"BO": "玻利维亚",
	"BA": "波斯尼亚和黑塞哥维那",
	"BW": "博茨瓦纳",
	"BR": "巴西",
	"BN": "文莱",
	"BG": "保加利亚",
	"BF": "布基纳法索",
	"BI": "布隆迪",
	"CV": "佛得角",
	"KH": "柬埔寨",
	"CM": "喀麦隆",
	"CA": "加拿大",
	"CF": "中非共和国",
	"TD": "乍得",
	"CL": "智利",
	"CN": "中国",
	"CO": "哥伦比亚",
	"KM": "科摩罗",
	"CG": "刚果",
	"CR": "哥斯达黎加",
	"HR": "克罗地亚",
	"CU": "古巴",
	"CY": "塞浦路斯",
	"CZ": "捷克共和国",
	"DK": "丹麦",
	"DJ": "吉布提",
	"DM": "多米尼克",
	"DO": "多米尼加共和国",
	"EC": "厄瓜多尔",
	"EG": "埃及",
	"SV": "萨尔瓦多",
	"GQ": "赤道几内亚",
	"ER": "厄立特里亚",
	"EE": "爱沙尼亚",
	"SZ": "斯威士兰",
	"ET": "埃塞俄比亚",
	"FJ": "斐济",
	"FI": "芬兰",
	"FR": "法国",
	"GA": "加蓬",
	"GM": "冈比亚",
	"GE": "格鲁吉亚",
	"DE": "德国",
	"GH": "加纳",
	"GR": "希腊",
	"GD": "格林纳达",
	"GT": "危地马拉",
	"GN": "几内亚",
	"GW": "几内亚比绍",
	"GY": "圭亚那",
	"HT": "海地",
	"HN": "洪都拉斯",
	"HU": "匈牙利",
	"IS": "冰岛",
	"IN": "印度",
	"ID": "印度尼西亚",
	"IR": "伊朗",
	"IQ": "伊拉克",
	"IE": "爱尔兰",
	"IL": "以色列",
	"IT": "意大利",
	"JM": "牙买加",
	"JP": "日本",
	"JO": "约旦",
	"KZ": "哈萨克斯坦",
	"KE": "肯尼亚",
	"KI": "基里巴斯",
	"KR": "韩国",
	"KP": "朝鲜",
	"KW": "科威特",
	"KG": "吉尔吉斯斯坦",
	"LA": "老挝",
	"LV": "拉脱维亚",
	"LB": "黎巴嫩",
	"LS": "莱索托",
	"LR": "利比里亚",
	"LY": "利比亚",
	"LI": "列支敦士登",
	"LT": "立陶宛",
	"LU": "卢森堡",
	"MG": "马达加斯加",
	"MW": "马拉维",
	"MY": "马来西亚",
	"MV": "马尔代夫",
	"ML": "马里",
	"MT": "马耳他",
	"MH": "马绍尔群岛",
	"MR": "毛里塔尼亚",
	"MU": "毛里求斯",
	"MX": "墨西哥",
	"FM": "密克罗尼西亚",
	"MD": "摩尔多瓦",
	"MC": "摩纳哥",
	"MN": "蒙古",
	"ME": "黑山",
	"MA": "摩洛哥",
	"MZ": "莫桑比克",
	"MM": "缅甸",
	"NA": "纳米比亚",
	"NR": "瑙鲁",
	"NP": "尼泊尔",
	"NL": "荷兰",
	"NZ": "新西兰",
	"NI": "尼加拉瓜",
	"NE": "尼日尔",
	"NG": "尼日利亚",
	"MK": "北马其顿",
	"NO": "挪威",
	"OM": "阿曼",
	"PK": "巴基斯坦",
	"PW": "帕劳",
	"PA": "巴拿马",
	"PG": "巴布亚新几内亚",
	"PY": "巴拉圭",
	"PE": "秘鲁",
	"PH": "菲律宾",
	"PL": "波兰",
	"PT": "葡萄牙",
	"QA": "卡塔尔",
	"RO": "罗马尼亚",
	"RU": "俄罗斯",
	"RW": "卢旺达",
	"KN": "圣基茨和尼维斯",
	"LC": "圣卢西亚",
	"VC": "圣文森特和格林纳丁斯",
	"WS": "萨摩亚",
	"SM": "圣马力诺",
	"ST": "圣多美和普林西比",
	"SA": "沙特阿拉伯",
	"SN": "塞内加尔",
	"RS": "塞尔维亚",
	"SC": "塞舌尔",
	"SL": "塞拉利昂",
	"SG": "新加坡",
	"SK": "斯洛伐克",
	"SI": "斯洛文尼亚",
	"SB": "所罗门群岛",
	"SO": "索马里",
	"ZA": "南非",
	"SS": "南苏丹",
	"ES": "西班牙",
	"LK": "斯里兰卡",
	"SD": "苏丹",
	"SR": "苏里南",
	"SE": "瑞典",
	"CH": "瑞士",
	"SY": "叙利亚",
	"TW": "台湾",
	"TJ": "塔吉克斯坦",
	"TZ": "坦桑尼亚",
	"TH": "泰国",
	"TL": "东帝汶",
	"TG": "多哥",
	"TO": "汤加",
	"TT": "特立尼达和多巴哥",
	"TN": "突尼斯",
	"TR": "土耳其",
	"TM": "土库曼斯坦",
	"TV": "图瓦卢",
	"UG": "乌干达",
	"UA": "乌克兰",
	"AE": "阿联酋",
	"GB": "英国",
	"US": "美国",
	"UY": "乌拉圭",
	"UZ": "乌兹别克斯坦",
	"VU": "瓦努阿图",
	"VA": "梵蒂冈",
	"VE": "委内瑞拉",
	"VN": "越南",
	"YE": "也门",
	"ZM": "赞比亚",
	"ZW": "津巴布韦",
	"HK": "香港",
}

// GetChineseCountryName 获取国家的中文名称
func GetChineseCountryName(englishName string) string {
	if chineseName, ok := CountryNameMap[englishName]; ok {
		return chineseName
	}
	return englishName // 如果找不到对应的中文名称，返回原始英文名称
}

// GetChineseCountryNameByCode 获取国家的中文名称
func GetChineseCountryNameByCode(countryCode string) string {
	if chineseName, ok := CountryCodeMap[countryCode]; ok {
		return chineseName
	}
	return countryCode // 如果找不到对应的中文名称，返回原始英文名称
}
