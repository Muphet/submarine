﻿using System;
using System.Collections;
using System.Collections.Generic;
using System.Reflection;
using System.Text;
using MiniJSON;

namespace TyphenApi
{
    public class JSONSerializer : ISerializer
    {
        public byte[] Serialize(object obj)
        {
            string text = string.Empty;

            if (obj != null && IsSerializableObject(obj.GetType()))
            {
                var dict = SerializeClassOrStruct(obj);
                text = Json.Serialize(dict);
            }

            return Encoding.UTF8.GetBytes(text);
        }

        public T Deserialize<T>(byte[] bytes) where T : new()
        {
            if (bytes.Length > 0)
            {
                var text = Encoding.UTF8.GetString(bytes);
                var dict = Json.Deserialize(text) as Dictionary<string, object>;
                var deserialized = DeserializeClassOrStruct(typeof(T), dict);
                return (T)deserialized;
            }
            else
            {
                return default(T);
            }
        }

        Dictionary<string, object> SerializeClassOrStruct(object obj)
        {
            var dict = new Dictionary<string, object>();
            var objType = obj.GetType();

            foreach (var property in objType.GetProperties())
            {
                var attributes = property.GetCustomAttributes(typeof(SerializablePropertyAttribute), true);

                foreach (var attribute in attributes)
                {
                    var metaInfo = (SerializablePropertyAttribute)attribute;
                    var value = property.GetValue(obj, null);

                    if (metaInfo.IsOptional && value == null)
                    {
                        continue;
                    }

                    var valueType = property.PropertyType;

                    if (value == null && IsNullableType(valueType))
                    {
                        var message = string.Format("{0}.{1} is not allowed to be null.", objType.FullName, property.Name);
                        throw new NoNullAllowedException(message);
                    }
                    else if (IsSerializableValue(value, valueType))
                    {
                        dict[metaInfo.PropertyName] = valueType.IsEnum ? (int)value : value;
                    }
                    else if (IsSerializableObject(valueType))
                    {
                        dict[metaInfo.PropertyName] = SerializeClassOrStruct(value);
                    }
                    else
                    {
                        var message = string.Format("Failed to serialize {0} ({1}) to {2}.{3}", value, valueType, objType.FullName, property.Name);
                        throw new SerializeFailedError(message);
                    }
                }
            }
            return dict;
        }

        object DeserializeClassOrStruct(System.Type objType, Dictionary<string, object> dict)
        {
            object obj = Activator.CreateInstance(objType);

            foreach (var property in objType.GetProperties())
            {
                var attributes = property.GetCustomAttributes(typeof(SerializablePropertyAttribute), true);

                foreach (var attribute in attributes)
                {
                    var metaInfo = (SerializablePropertyAttribute)attribute;
                    var value = dict[metaInfo.PropertyName];

                    if (value == null)
                    {
                        if (metaInfo.IsOptional)
                        {
                            continue;
                        }
                        else
                        {
                            var message = string.Format("{0}.{1} is not allowed to be null.", objType.FullName, property.Name);
                            throw new NoNullAllowedException(message);
                        }
                    }

                    var valueType = property.PropertyType;

                    if (IsSerializableValue(value, valueType))
                    {
                        property.SetValue(obj, Convert.ChangeType(value, valueType), null);
                    }
                    else if (IsSerializableObject(valueType))
                    {
                        var parsed = DeserializeClassOrStruct(valueType, (Dictionary<string, object>)value);
                        property.SetValue(obj, parsed, null);
                    }
                    else
                    {
                        var message = string.Format("Failed to deserialize {0} ({1}) to {2}.{3}", value, valueType, objType.FullName, property.Name);
                        throw new DeserializeFailedError(message);
                    }
                }
            }
            return obj;
        }

        bool IsNullableType(System.Type type)
        {
            return type.IsClass || Nullable.GetUnderlyingType(type) != null;
        }

        bool IsIDictionaryImplementation(System.Type type)
        {
            return typeof(IDictionary).IsAssignableFrom(type);
        }

        bool IsStruct(System.Type type)
        {
            return type.IsValueType && !type.IsEnum;
        }

        bool IsSerializableObject(System.Type type)
        {
            return type.IsClass || IsStruct(type);
        }

        bool IsSerializableValue(object value, System.Type valueType)
        {
            return valueType.IsPrimitive || valueType.IsEnum ||
                value is string || value is IList ||
                (value is IDictionary && IsIDictionaryImplementation(valueType));
        }
    }
}