﻿using System;

namespace TyphenApi
{
    [AttributeUsage(AttributeTargets.Property, AllowMultiple = false, Inherited = true)]
    public class SerializablePropertyAttribute : Attribute
    {
        public string PropertyName { get; private set; }
        public bool IsOptional { get; private set; }

        public SerializablePropertyAttribute(string propertyName, bool isOptional)
        {
            PropertyName = propertyName;
            IsOptional = isOptional;
        }
    }
}