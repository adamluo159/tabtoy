# Tabtoy V2 使用说明书

## 目录

1. [表格结构](#表格结构)
2. [@Types 类型定义表](#types-类型定义表)
3. [Meta 标签说明](#meta-标签说明)
4. [Alias 别名功能](#alias-别名功能)
5. [数据表内定义枚举 (StandKey/StandAlias)](#数据表内定义枚举-standkeystandalias)
6. [Map 类型使用](#map-类型使用)
7. [常见问题](#常见问题)

---

## 表格结构

每个 Excel 文件包含两种类型的 Sheet：

### 1. @Types 类型定义表（必须存在）

用于定义枚举类型和结构体类型。

### 2. 数据表

实际的数据内容，可以有多个 Sheet。

### 数据表行结构

| 行号 | 内容 |
|------|------|
| 第0行 | 字段名 |
| 第1行 | 字段类型 |
| 第2行 | Meta（特性） |
| 第3行 | 注释 |
| 第4行开始 | 数据 |

---

## @Types 类型定义表

### 表格结构

| 行号 | 内容 |
|------|------|
| 第0行 | 配置信息，如 `TableName: "Item" Package: "config"` |
| 第1行 | 字段描述头（ObjectType, FieldName, FieldType, Value, Alias, Default, Meta, Comment） |
| 第2行 | 注释行 |
| 第3行开始 | 类型定义数据 |

### 定义枚举类型

```
TableName: "Globals" Package: "table"
ObjectType  FieldName   FieldType   Value   Alias   Default   Meta                        Comment
对象类型     字段名       字段类型     枚举值   别名     默认值     特性                        注释
ActorType   Leader      int32       0       唐僧               LuaValueMapperString:true   Lua枚举的映射方式
ActorType   Monkey      int32       1       孙悟空             LuaStringMapperValue:true
ActorType   Pig         int32       2       猪八戒             LuaValueMapperString:true
ActorType   Hammer      int32       3       沙僧               LuaStringMapperValue:true
```

### 定义结构体类型

```
ObjectType       FieldName   FieldType   Value   Alias   Default   Meta            Comment
BulletUpGradeCost Bullet     BulletType                            子弹类型
BulletUpGradeCost Coin       int32                                  资源数
```

---

## Meta 标签说明

Meta 列用于添加字段特性，多个标签用空格分隔。

### 系统内置标签

| 标签 | 说明 | 示例 |
|------|------|------|
| **MakeIndex** | 创建索引，用于快速查找 | `MakeIndex:true` |
| **RepeatCheck** | 重复检查，确保值唯一 | `RepeatCheck:true` |
| **MustFill** | 必填字段 | `MustFill:true` |
| **Default** | 默认值 | `Default:0` |
| **ListSpliter** | 列表分隔符 | `ListSpliter:\|` |
| **MapSpliter** | Map 分隔符 | `MapSpliter:\|` |
| **MapKeyField** | Map 的 Key 字段名 | `MapKeyField:ID` |
| **Mark** | 标记字段，用于筛选输出 | `Mark:Client` |
| **SimpleInput** | 简化输入，用于2字段结构体 | `SimpleInput:true` |

### Lua 相关标签

| 标签 | 说明 |
|------|------|
| **LuaValueMapperString** | Lua 枚举值映射为字符串 |
| **LuaStringMapperValue** | Lua 字符串映射为枚举值 |

### 数据表内定义枚举标签

| 标签 | 必填 | 说明 |
|------|------|------|
| **StandKey** | ✅ 是 | 枚举的数值列，必须是整数 |
| **StandCode** | ❌ 否 | 枚举的英文名（字段名） |
| **StandAlias** | ✅ 是 | 枚举的中文别名 |
| **StandName** | ❌ 否 | 枚举的注释/名称 |

### 用户自定义标签

除了系统内置标签，可以添加任意自定义标签，这些标签会作为 struct tag 输出到生成的代码中。

示例：
```
Meta: `MyTag:"value" JsonTag:"my_field"`
```

生成的 Go 代码：
```go
FieldName string `MyTag:"value" JsonTag:"my_field"`
```

---

## SimpleInput 简化输入

### 用途

当结构体只有 2 个字段时，可以使用简化格式输入，减少输入字符数。

### 使用方法

在 @Types 表中，给结构体的**第一个字段**添加 `SimpleInput:true` 标签：

```
ObjectType    FieldName   FieldType   Value   Meta
ItemCost      ID          int32               SimpleInput:true
ItemCost      Num         int32
```

### 输入格式对比

**标准格式**（需要写字段名）：
```
ID:金币 Num:2000|ID:一阶碎片 Num:40
```

**简化格式**（不需要写字段名）：
```
金币:2000|一阶碎片:40
```

### 完整示例

**@Types 表定义：**

| ObjectType | FieldName | FieldType | Value | Meta |
|------------|-----------|-----------|-------|------|
| ItemCost | ID | int32 | | `SimpleInput:true` |
| ItemCost | Num | int32 | | |

**数据表中使用：**

| 字段名 | 类型 | Meta | 值 |
|--------|------|------|-----|
| Costs | `repeated ItemCost` | `ListSpliter:\|` | `金币:2000\|一阶碎片:40` |

**等价于标准格式：**

| 字段名 | 类型 | Meta | 值 |
|--------|------|------|-----|
| Costs | `repeated ItemCost` | `ListSpliter:\|` | `ID:金币 Num:2000\|ID:一阶碎片 Num:40` |

### 注意事项

1. 结构体必须**恰好有 2 个字段**
2. `SimpleInput:true` 标签必须添加在**第一个字段**上
3. 简化格式为 `值1:值2`，按字段定义顺序对应
4. 如果简化格式解析失败，会自动尝试标准格式

---

## Alias 别名功能

### 用途

Alias 用于给枚举值设置中文别名，使得在数据表中填写枚举值时可以使用中文别名代替英文名。

### 使用方法

在 @Types 表的 Alias 列填写中文别名：

| ObjectType | FieldName | FieldType | Value | **Alias** |
|------------|-----------|-----------|-------|-----------|
| ActorType | Leader | int32 | 0 | **唐僧** |
| ActorType | Monkey | int32 | 1 | **孙悟空** |
| ActorType | Pig | int32 | 2 | **猪八戒** |

### 数据表中使用

定义了别名后，在数据表中可以直接使用别名：

| ID | ActorType |
|----|-----------|
| 1 | 唐僧 | ← 使用别名 |
| 2 | Leader | ← 使用字段名 |
| 3 | 孙悟空 | ← 使用别名 |

三种写法都能正确解析为对应的枚举值。

---

## 数据表内定义枚举 (StandKey/StandAlias)

### 用途

在数据表中直接定义枚举类型，而不需要在 @Types 表中单独定义。适用于枚举值和数据紧密相关的场景。

### 使用方法

在数据表的 Meta 行添加标签：

| | A列(ID) | B列(Name) | C列(Type) |
|--|---------|-----------|-----------|
| **第0行** | ID | Name | Type |
| **第1行** | int32 | string | int32 |
| **第2行** | `StandKey:true` | `StandAlias:true` | |
| **第3行** | 道具ID | 名称 | 类型 |
| **第4行** | 10000 | 金币 | 1 |
| **第5行** | 10100 | 钻石 | 2 |

### 标签说明

| 标签 | 必填 | 说明 |
|------|------|------|
| **StandKey** | ✅ 是 | 枚举的数值列，必须是整数 |
| **StandCode** | ❌ 否 | 枚举的英文名，如果没有则不输出枚举代码 |
| **StandAlias** | ✅ 是 | 枚举的中文别名 |
| **StandName** | ❌ 否 | 枚举的注释/名称 |

### 生成的枚举名

枚举名 = Sheet名 + 字段名

例如：Sheet名为 "Item"，字段名为 "ID"，则生成的枚举名为 `ItemID`

### 示例

**Item 表：**

| ID | Name |
|----|------|
| int32 | string |
| `StandKey:true` | `StandAlias:true` |
| 道具ID | 名称 |
| 10000 | 金币 |
| 10100 | 钻石 |
| 10200 | 体力 |

**生成的枚举：**

```go
type ItemID int32

const (
    ItemIDNone   ItemID = 0    // 自动生成
    ItemID金币   ItemID = 10000
    ItemID钻石   ItemID = 10100
    ItemID体力   ItemID = 10200
)
```

### 注意事项

1. 如果没有 `StandCode` 列，则枚举的 `NotPrint=true`，不会输出到代码中，但仍可在其他表中使用该类型
2. `StandKey` 和 `StandAlias` 必须同时存在才能生效

---

## Map 类型使用

### 定义 Map 字段

在数据表的类型行使用 `map<K,V>` 格式：

| 字段名 | 类型 | Meta |
|--------|------|------|
| CostMap | `map<int32,int32>` | `MapSpliter:"\|"` |

### 填写数据

使用分隔符分隔多个键值对：

```
1:10|2:20|3:30
```

### 使用枚举作为 Key

```
map<BulletType,int32>
```

### 使用结构体作为 Value

```
map<int32,ItemCost>
```

需要配合 `MapKeyField` 指定结构体中作为 Key 的字段：

| 字段名 | 类型 | Meta |
|--------|------|------|
| BulletMap | `map<BulletType,BulletAbilityDesc>` | `MapSpliter:"\|" MapKeyField:Bullet` |

---

## 常见问题

### Q: 提示 "Type not found" 错误

**原因**：引用的类型没有定义。

**解决方案**：
1. 检查 @Types 表中是否定义了该类型
2. 如果使用 StandKey/StandAlias 定义枚举，确保文件处理顺序正确（新版本已支持任意顺序）
3. 检查类型名拼写是否正确

### Q: 枚举值必须从 0 开始吗？

是的，protobuf3 要求枚举必须有 0 值。如果不定义 0 值，会报错。

### Q: 如何在多个表中共享类型？

在任意一个文件的 @Types 表中定义类型，该类型会被添加到全局类型描述符中，其他文件可以直接引用。

### Q: StandKey 定义的枚举为什么没有输出到代码？

如果没有 `StandCode` 列，枚举的 `NotPrint=true`，不会输出到代码中。如果需要输出，请添加 `StandCode` 列。

### Q: 如何筛选输出字段？

使用 `Mark` 标签标记字段，然后在导出时指定 `--field_mark` 参数。

```
Meta: `Mark:Client`
```

只输出标记为 Client 的字段。

---

## 文件处理流程

新版本采用两阶段解析，解决了类型依赖顺序问题：

```
第一阶段：收集所有类型
├── 解析所有文件的 @Types 表
├── 解析所有文件的 StandKey/StandAlias 定义
└── 将所有类型添加到全局 FileDescriptor

第二阶段：解析所有表头
├── 此时所有类型都已可用
└── 解析数据表头，建立字段类型引用
```

这意味着无论文件顺序如何，都可以正确解析类型引用。
